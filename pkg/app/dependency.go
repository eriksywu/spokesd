package app

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/eriksywu/spokesd/pkg/controller"
	"github.com/eriksywu/spokesd/pkg/k8s"
)

var Handlers = make(map[string]func())

// App should hold all Watchers, configs and expose a channels that will be multiplexed and/or fanned to all watchers
type App struct {
	watcher *controller.Watcher
}

// this is dumb
var resources = []runtime.Object{
	//&v1.Node{TypeMeta: metav1.TypeMeta{Kind: "nodes"}},
	&v1.Pod{TypeMeta: metav1.TypeMeta{Kind: "pods"}},
}

func NewApp() (*App, error) {
	client, err := k8s.GetClientFromEnv()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resourcesToWatch := make(map[runtime.Object]*controller.ResourceWatchConfig)
	for _, resource := range resources {
		resourcesToWatch[resource] = controller.NewResourceWatchConfig(resource)
	}

	watcher := &controller.Watcher{Resources: resourcesToWatch, ClientSet: client}
	return &App{watcher}, nil
}

func (a *App) Run(stop chan struct{}) {
	a.watcher.RunAsyncWithContext(context.Background(), stop)
}
