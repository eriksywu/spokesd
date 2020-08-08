package controller

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/util/workqueue"

	"k8s.io/client-go/informers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/eriksywu/spokesd/pkg/model"
)

type opt func(options *metav1.ListOptions)

type QueueFactory func() workqueue.RateLimitingInterface

// NewInformerFunc takes kubernetes.Interface and time.Duration to return a SharedIndexInformer.
type NewInformerFunc func(kubernetes.Interface, time.Duration) cache.SharedIndexInformer

type DeleteFnFactory func(workqueue.RateLimitingInterface, ...MutatingHook) func(obj interface{})
type AddFnFactory func(workqueue.RateLimitingInterface, ...MutatingHook) func(obj interface{})
type UpdateFnFactory func(workqueue.RateLimitingInterface, ...MutatingHook) func(oldObj, newObj interface{})

type ResourceWatchConfig struct {
	OptionsFns           []opt
	Kind                 schema.GroupVersionKind
	AddFnFactory         AddFnFactory
	DeleteFnFactory      DeleteFnFactory
	UpdateFnFactory      UpdateFnFactory
	restartWatchOnDelete bool
	watch                *ResourceWatch

	resyncPeriod time.Duration
}

func NewResourceWatchConfig(obj runtime.Object) *ResourceWatchConfig {
	config := &ResourceWatchConfig{}
	config.AddFnFactory = DefaultAddFuncFactory
	config.DeleteFnFactory = DefaultDeleteFuncFactory
	config.UpdateFnFactory = DefaultUpdateFunFactory
	config.Kind = obj.GetObjectKind().GroupVersionKind()
	return config
}

func (c *ResourceWatchConfig) GetWatchController(clientSet k8s.Interface, obj runtime.Object, factory informers.SharedInformerFactory,
	qFactory QueueFactory, opts ...opt) (*ResourceWatch, error) {

	if c.watch != nil {
		return c.watch, nil
	}
	clientSet.CoreV1()

	lw := c.newListWatch(clientSet, opts...)

	resourceWatch := &ResourceWatch{}
	if qFactory == nil {
		resourceWatch.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	} else {
		resourceWatch.queue = qFactory()
	}

	handler := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.AddFnFactory(resourceWatch.queue),
		DeleteFunc: c.DeleteFnFactory(resourceWatch.queue),
		UpdateFunc: c.UpdateFnFactory(resourceWatch.queue),
	}

	resourceWatch.informer = factory.InformerFor(obj, c.NewDefaultInformer(obj, lw))
	resourceWatch.informer.AddEventHandler(handler)
	c.watch = resourceWatch
	return resourceWatch, nil
}

func (c *ResourceWatchConfig) newListWatch(clientSet kubernetes.Interface, opts ...opt) *cache.ListWatch {
	optionFn := func(options *metav1.ListOptions) {
		for _, fn := range c.OptionsFns {
			fn(options)
		}
		for _, fn := range opts {
			fn(options)
		}
	}
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			optionFn(&options)
			return clientSet.CoreV1().RESTClient().Get().Resource(c.Kind.Kind).
				VersionedParams(&options, scheme.ParameterCodec).
				Do().Get()
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			optionFn(&options)
			options.Watch = true
			return clientSet.CoreV1().RESTClient().Get().Resource(c.Kind.Kind).
				VersionedParams(&options, scheme.ParameterCodec).Watch()
		},
	}
	return lw
}

func (c *ResourceWatchConfig) NewDefaultInformer(obj runtime.Object, lw *cache.ListWatch) func(kubernetes.Interface, time.Duration) cache.SharedIndexInformer {
	return func(kubernetes.Interface, time.Duration) cache.SharedIndexInformer {
		return cache.NewSharedIndexInformer(
			lw,
			obj,
			c.resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)
	}
}

type ResourceWatch struct {
	informer cache.SharedInformer
	queue    workqueue.RateLimitingInterface
}

func (rw *ResourceWatch) Worker() {
	for rw.processItem() {
	}
}

func (rw *ResourceWatch) processItem() bool {
	obj, quit := rw.queue.Get()
	if quit {
		return false
	}
	defer func() {
		rw.queue.Forget(obj)
		rw.queue.Done(obj)
	}()

	event, k := obj.(*model.Event)
	if !k {
		return false
	}

	fmt.Printf("processing event %v in queue \n", event)
	return true
}
