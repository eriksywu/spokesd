package model

import "k8s.io/apimachinery/pkg/runtime/schema"

type T interface{}

type Type string

const (
	Update Type = "Update"
	Add    Type = "Add"
	Delete Type = "Delete"
)

type Event struct {
	QueueKey string
	Kind     schema.GroupVersionKind
	Type     Type
	Data     T
}

func (e *Event) T() {

}
