package controller

import (
	"github.com/eriksywu/spokesd/pkg/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type MutatingHook func(*model.Event)

// default Func factories
func DefaultAddFuncFactory(queue workqueue.RateLimitingInterface, hooks ...MutatingHook) func(obj interface{}) {
	return func(obj interface{}) {
		k8sObj, k := obj.(runtime.Object)
		if !k {
			return
		}
		event := &model.Event{
			Kind: k8sObj.GetObjectKind().GroupVersionKind(),
			Type: model.Add,
		}
		event.QueueKey, _ = cache.MetaNamespaceKeyFunc(obj)
		for _, h := range hooks {
			h(event)
		}
		queue.Add(event)
	}
}

func DefaultDeleteFuncFactory(queue workqueue.RateLimitingInterface, hooks ...MutatingHook) func(obj interface{}) {
	return func(obj interface{}) {
		k8sObj, k := obj.(runtime.Object)
		if !k {
			return
		}
		event := &model.Event{
			Kind: k8sObj.GetObjectKind().GroupVersionKind(),
			Type: model.Delete,
		}
		event.QueueKey, _ = cache.MetaNamespaceKeyFunc(obj)
		for _, h := range hooks {
			h(event)
		}
		queue.Add(event)
	}
}

func DefaultUpdateFunFactory(queue workqueue.RateLimitingInterface, hooks ...MutatingHook) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		k8sObj, k := newObj.(runtime.Object)
		if !k {
			return
		}
		event := &model.Event{
			Kind: k8sObj.GetObjectKind().GroupVersionKind(),
			Type: model.Update,
		}
		event.QueueKey, _ = cache.MetaNamespaceKeyFunc(newObj)
		event.Data = oldObj
		for _, h := range hooks {
			h(event)
		}
		queue.Add(event)
	}
}
