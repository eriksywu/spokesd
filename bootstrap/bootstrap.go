package boostrap

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type BootstrapFn func(object runtime.Object)

func init() {

}
