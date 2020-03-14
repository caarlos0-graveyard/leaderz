package main

import (
	"context"
	"flag"
	"sync/atomic"
	"time"

	"github.com/apex/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

func main() {
	var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	var nodeID = flag.String("node-id", "", "node id used for leader election")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientset, err := newClientset(*kubeconfig)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to cluster")
	}

	var lock = &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "my-lock",
			Namespace: "default",
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: *nodeID,
		},
	}

	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()
	var leading int32
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				atomic.StoreInt32(&leading, 1)
				log.WithField("id", *nodeID).Info("started leading")
				for range ticker.C {
					if atomic.LoadInt32(&leading) == 0 {
						log.Info("stopped working")
						return
					}
					log.Info("working...")
					time.Sleep(time.Second)
				}
			},
			OnStoppedLeading: func() {
				atomic.StoreInt32(&leading, 0)
				log.WithField("id", *nodeID).Info("stopped leading")
			},
			OnNewLeader: func(identity string) {
				if identity == *nodeID {
					return
				}
				log.WithField("id", *nodeID).
					WithField("leader", identity).
					Info("new leader")
			},
		},
	})
}

func newClientset(filename string) (*kubernetes.Clientset, error) {
	config, err := getConfig(filename)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getConfig(cfg string) (*rest.Config, error) {
	if cfg == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", cfg)
}
