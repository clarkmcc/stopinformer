package stopinformermap

import (
	"github.com/clarkmcc/stopinformer/goroutinemap"
	"github.com/clarkmcc/stopinformer/stopinformer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewGenericStopInformerMap(t *testing.T) {
	informers := NewGenericStopInformerMap()

	informers.Create("informer1", stopinformer.NewGenericStopInformer().ResolveImmediately())
	informers.Create("informer2", stopinformer.NewGenericStopInformer().ResolveImmediately())

	informers.StopAll()

	assert.Equal(t, false, informers.Exists("informer1"))
	assert.Equal(t, false, informers.Exists("informer2"))
}

func TestNewGenericStopInformerMap2(t *testing.T) {
	informers := NewGenericStopInformerMap()

	informer1 := informers.CreateAndReturn("informer1", stopinformer.NewGenericStopInformer())
	informer2 := informers.CreateAndReturn("informer2", stopinformer.NewGenericStopInformer())

	routineMap := goroutinemap.NewGoRoutineMap(false)
	routineMap.Run("routine1", func() error {
	loop:
		for {
			select {
			case stop := <-informer1.Watch():
				defer stop.Acknowledge()
				break loop
			case <-time.After(time.Second):
				continue
			}
		}
		return nil
	})
	routineMap.Run("routine2", func() error {
	loop:
		for {
			select {
			case stop := <-informer2.Watch():
				defer stop.Acknowledge()
				break loop
			case <-time.After(time.Second):
				continue
			}
		}
		return nil
	})

	assert.Equal(t, true, routineMap.IsOperationPending("routine1"))
	assert.Equal(t, true, routineMap.IsOperationPending("routine2"))
	assert.Equal(t, true, informers.Exists("informer1"))
	assert.Equal(t, true, informers.Exists("informer2"))

	<-informers.StopAllAndNotify(0)

	assert.Equal(t, false, routineMap.IsOperationPending("routine1"))
	assert.Equal(t, false, routineMap.IsOperationPending("routine2"))
	assert.Equal(t, false, informers.Exists("informer1"))
	assert.Equal(t, false, informers.Exists("informer2"))
}
