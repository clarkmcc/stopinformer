package stopinformers

import (
	"github.com/clarkmcc/stopinformer/goroutinemap"
	"github.com/clarkmcc/stopinformer/stopinformer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewGenericStopInformers(t *testing.T) {
	informers := NewGenericStopInformers()

	informers.Create(stopinformer.NewGenericStopInformer().ResolveImmediately())
	informers.Create(stopinformer.NewGenericStopInformer().ResolveImmediately())

	informers.StopAll()
}

func TestNewGenericStopInformers2(t *testing.T) {
	informers := NewGenericStopInformers()

	informer1 := informers.CreateAndReturn(stopinformer.NewGenericStopInformer())
	informer2 := informers.CreateAndReturn(stopinformer.NewGenericStopInformer())

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

	<-informers.StopAllAndNotify(0)

	assert.Equal(t, false, routineMap.IsOperationPending("routine1"))
	assert.Equal(t, false, routineMap.IsOperationPending("routine2"))
}
