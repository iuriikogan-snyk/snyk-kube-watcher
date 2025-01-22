package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/config"
	"github.com/iuriikogan-snyk/snyk-kube-watcher/internal/tasks"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func StartWatchers(ctx context.Context, c config.Config, taskCh chan<- tasks.ImageTask) *sync.WaitGroup {
	var wg sync.WaitGroup
	for _, cl := range c.Clusters {
		wg.Add(1)
		go func(cfg config.ClusterConfig) {
			defer wg.Done()
			err := runInformer(ctx, cfg, c.OrgID, taskCh)
			if err != nil {
				slog.Error("Watcher error", "kubeconfig", cfg.Kubeconfig, "context", cfg.Context, "err", err)
			}
		}(cl)
	}
	return &wg
}

func runInformer(ctx context.Context, cl config.ClusterConfig, orgID string, taskCh chan<- tasks.ImageTask) error {
	slog.Info("Watcher start", "kubeconfig", cl.Kubeconfig, "context", cl.Context)
	restCfg, err := buildConfig(cl.Kubeconfig, cl.Context)
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return err
	}
	infFactory := informers.NewSharedInformerFactory(client, 10*time.Minute)
	podInf := infFactory.Core().V1().Pods().Informer()
	podInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			enqueue(obj, orgID, taskCh)
		},
		UpdateFunc: func(_, newObj interface{}) {
			enqueue(newObj, orgID, taskCh)
		},
	})
	stopCh := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(stopCh)
	}()
	infFactory.Start(stopCh)
	for t, ok := range infFactory.WaitForCacheSync(stopCh) {
		if !ok {
			return fmt.Errorf("sync failed for %s", t.String())
		}
	}
	<-stopCh
	slog.Info("Watcher stop", "kubeconfig", cl.Kubeconfig, "context", cl.Context)
	return nil
}

func enqueue(obj interface{}, orgID string, taskCh chan<- tasks.ImageTask) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	for _, c := range pod.Spec.Containers {
		taskCh <- tasks.ImageTask{Image: c.Image, OrgID: orgID}
	}
}

func buildConfig(kubeconfigPath, contextName string) (*rest.Config, error) {
	rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	return cfg.ClientConfig()
}
