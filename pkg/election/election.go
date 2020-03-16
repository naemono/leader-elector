package election

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typed_core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	election "k8s.io/client-go/tools/leaderelection"
	election_resourcelock "k8s.io/client-go/tools/leaderelection/resourcelock"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

func getCurrentLeader(electionId, namespace string, c kubernetes.Interface) (string, *corev1.Endpoints, error) {
	endpoints, err := c.CoreV1().Endpoints(namespace).Get(electionId, metav1.GetOptions{})
	if err != nil {
		return "", nil, err
	}
	val, found := endpoints.Annotations[election_resourcelock.LeaderElectionRecordAnnotationKey]
	if !found {
		return "", endpoints, nil
	}
	electionRecord := election_resourcelock.LeaderElectionRecord{}
	if err := json.Unmarshal([]byte(val), &electionRecord); err != nil {
		return "", nil, err
	}
	return electionRecord.HolderIdentity, endpoints, err
}

func New(id, name, namespace string, ttl time.Duration, callback func(leader string), kubeClient kubernetes.Interface) (*election.LeaderElector, error) {
	var (
		err      error
		hostname string
	)
	objectMeta := metav1.ObjectMeta{Namespace: namespace, Name: name}
	hostname, err = os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get hostname")
	}
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&typed_core_v1.EventSinkImpl{
			Interface: kubeClient.CoreV1().Events(namespace)})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		corev1.EventSource{
			Component: "leader-elector",
			Host:      hostname,
		})
	resourceLockConfig := rl.ResourceLockConfig{
		Identity:      name,
		EventRecorder: recorder,
	}
	lock := &rl.EndpointsLock{
		EndpointsMeta: objectMeta,
		LockConfig:    resourceLockConfig,
		Client:        kubeClient.CoreV1(),
	}
	callbacks := election.LeaderCallbacks{
		OnStartedLeading: func(context.Context) {
			callback(id)
		},
		OnStoppedLeading: func() {
			leader, _, err := getCurrentLeader(id, namespace, kubeClient)
			if err != nil {
				glog.Errorf("failed to get leader: %v", err)
				// empty string means leader is unknown
				callback("")
				return
			}
			callback(leader)
		},
		OnNewLeader: func(identity string) {
			callback(identity)
		},
	}
	return election.NewLeaderElector(election.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod:   ttl / 4,
		Callbacks:     callbacks,
	})
}
