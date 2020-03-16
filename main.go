/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	pkg_election "github.com/naemono/leader-elector/pkg/election"
)

var (
	name      = flag.String("election", "", "The name of the election")
	namespace = flag.String("election-namespace", metav1.NamespaceDefault, "The Kubernetes namespace for this election")
	ttl       = flag.Duration("ttl", 10*time.Second, "The TTL for this election's lease duration")
	inCluster = flag.Bool("use-cluster-credentials", false, "Should this request use cluster credentials?")
	addr      = flag.String("http", "", "If non-empty, stand up a simple webserver that reports the leader state")

	leader = &LeaderData{}
)

func makeClient() (kubernetes.Interface, error) {
	var (
		client kubernetes.Interface
		config *rest.Config
		err    error
	)
	// creates the in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// LeaderData represents information about the current leader
type LeaderData struct {
	Name string `json:"name"`
}

func webHandler(res http.ResponseWriter, req *http.Request) {
	data, err := json.Marshal(leader)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(data)
}

func validateFlags() {
	if len(*name) == 0 {
		glog.Fatal("--election cannot be empty")
	}
}

func init() {
	flag.Parse()
}

func main() {
	validateFlags()

	kubeClient, err := makeClient()
	if err != nil {
		glog.Fatalf("error connecting to the client: %v", err)
	}

	fn := func(str string) {
		leader.Name = str
		fmt.Printf("%s is the leader\n", leader.Name)
	}

	e, err := pkg_election.New(*name, *namespace, *ttl, fn, kubeClient)
	if err != nil {
		glog.Fatalf("failed to create election: %v", err)
	}
	go e.Run(context.Background())

	if len(*addr) > 0 {
		http.HandleFunc("/", webHandler)
		http.ListenAndServe(*addr, nil)
	} else {
		select {}
	}
}
