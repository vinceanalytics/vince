package control

import (
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Op uint

const (
	ADD Op = iota
	Update
	Delete
)

type Work struct {
	Op   Op
	Item any
}

type enqueueWorkHandler struct {
	logger    *zerolog.Logger
	workQueue chan *Work
}

// OnAdd is called when an object is added to the informers cache.
func (h *enqueueWorkHandler) OnAdd(obj interface{}, ok bool) {
	h.workQueue <- &Work{
		Op: ADD, Item: obj,
	}
}

// OnUpdate is called when an object is updated in the informers cache.
func (h *enqueueWorkHandler) OnUpdate(oldObj interface{}, newObj interface{}) {
	oldObjMeta, okOld := oldObj.(metav1.Object)
	newObjMeta, okNew := newObj.(metav1.Object)
	// This is a resync event, no extra work is needed.
	if okOld && okNew && oldObjMeta.GetResourceVersion() == newObjMeta.GetResourceVersion() {
		return
	}
	h.workQueue <- &Work{
		Op: Update, Item: newObj,
	}
}

// OnDelete is called when an object is removed from the informers cache.
func (h *enqueueWorkHandler) OnDelete(obj interface{}) {
	h.workQueue <- &Work{
		Op: Delete, Item: obj,
	}
}
