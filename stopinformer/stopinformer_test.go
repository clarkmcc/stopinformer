package stopinformer

import (
	"sync"
	"testing"
	"time"
)

type routineTracker struct {
	routines map[int]struct{}
	m        *sync.Mutex
}

func TestNewGenericStopInformer(t *testing.T) {
	routineTracker := &routineTracker{
		routines: map[int]struct{}{},
		m:        &sync.Mutex{},
	}

	informer := NewGenericStopInformer()
	informer2 := NewGenericStopInformer()

	go func(id int) {
		routineTracker.m.Lock()
		routineTracker.routines[id] = struct{}{}
		routineTracker.m.Unlock()

		defer func() {
			routineTracker.m.Lock()
			defer routineTracker.m.Unlock()
			delete(routineTracker.routines, id)
		}()

	loop:
		for {
			select {
			case stop := <-informer.Watch():
				defer stop.Acknowledge()
				break loop
			case <-time.After(1 * time.Second):
				continue
			}
		}
	}(1)

	go func(id int) {
		routineTracker.m.Lock()
		routineTracker.routines[id] = struct{}{}
		routineTracker.m.Unlock()

		defer func() {
			routineTracker.m.Lock()
			defer routineTracker.m.Unlock()
			delete(routineTracker.routines, id)
		}()

	loop:
		for {
			select {
			case stop := <-informer2.Watch():
				defer stop.Acknowledge()
				break loop
			case <-time.After(1 * time.Second):
				continue
			}
		}
	}(2)

	time.Sleep(time.Millisecond)

	routineTracker.m.Lock()
	if _, ok := routineTracker.routines[1]; !ok {
		t.Fatalf("expected routine id 1 to be running")
	}
	routineTracker.m.Unlock()

	informer.Stop()
	time.Sleep(time.Millisecond)

	routineTracker.m.Lock()
	if _, ok := routineTracker.routines[1]; ok {
		t.Fatalf("expected routine id 1 to be stopped")
	}
	if _, ok := routineTracker.routines[2]; !ok {
		t.Fatalf("expected routine id 2 to be running")
	}
	routineTracker.m.Unlock()

	<-informer2.StopAndNotify(0)
	routineTracker.m.Lock()
	if _, ok := routineTracker.routines[2]; ok {
		t.Fatalf("expected routine id 2 to be stopped")
	}
	routineTracker.m.Unlock()
}
