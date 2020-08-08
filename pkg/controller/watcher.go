package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/informers"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8s "k8s.io/client-go/kubernetes"
)

// a watcher object encapsulates watcher/queue resources of an entire k8s cluster
// thus we will use an shared informer to optimize caching
type Watcher struct {
	ClientSet k8s.Interface
	//actions   map[string][]boostrap.BootstrapFn
	Resources map[runtime.Object]*ResourceWatchConfig

	cache cache.Store
}

type WOChannel struct {
	c chan<- struct{}
}

func (w *Watcher) watchContext(ctx context.Context, stops []WOChannel, stop <-chan struct{}) {
	select {
	case <-ctx.Done():
		for _, ch := range stops {
			ch.c <- struct{}{}
		}
	case <-stop:
		os.Exit(1)
	}
}

func (w *Watcher) RunAsyncWithContext(ctx context.Context, stop <-chan struct{}) chan<- struct{} {
	done := make(chan struct{})
	stopChannels := make([]WOChannel, 0, len(w.Resources))
	factory := informers.NewSharedInformerFactory(w.ClientSet, 0)
	// TODO: de-multiplex
	stopDemux := stop

	// for each object type, construct a watch on it
	for obj, opt := range w.Resources {
		resourceWatch, _ := opt.GetWatchController(w.ClientSet, obj, factory, nil)
		stopChan := make(chan struct{})
		stopChannels = append(stopChannels, WOChannel{stopChan})
		// TODO tidy this up
		go startWatch(resourceWatch, stopChan)
	}
	w.watchContext(ctx, stopChannels, stopDemux)
	return done
}

func startWatch(resourceWatch *ResourceWatch, stopCh <-chan struct{}) {
	defer resourceWatch.queue.ShutDown()
	go resourceWatch.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, resourceWatch.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	wait.Until(resourceWatch.Worker, time.Second, stopCh)
}
