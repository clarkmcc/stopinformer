package stopinformermap

import (
	"github.com/clarkmcc/stopinformer/stopinformer"
	"sync"
)

// StopInformerMap provides an easy interface for interacting with named stop informers
type StopInformerMap interface {
	// Starts a new named stop informer
	Create(name string, informer stopinformer.Interface)

	// Starts a new named stop informer and returns it for method chaining
	CreateAndReturn(name string, informer stopinformer.Interface) stopinformer.Interface

	// Returns a stop informer
	Get(name string) stopinformer.Interface

	// Returns whether a stop informer exists
	Exists(name string) bool

	// Iterates over all stop informers and sends the stop notification to each; this method
	// blocks until all informers acknowledge the stop command
	StopAll()

	// Iterates over all stop informers and sends the stop notification to each; this method
	// returns a channel and sends a value when all informers acknowledge they're stopped
	StopAllAndNotify(buffer int) stopinformer.SingleStructChan
}

type genericStopInformerMap struct {
	informers map[string]stopinformer.Interface
	m         *sync.RWMutex
}

func NewGenericStopInformerMap() StopInformerMap {
	return &genericStopInformerMap{
		informers: map[string]stopinformer.Interface{},
		m:         &sync.RWMutex{},
	}
}

// Starts a new named stop informer
func (g *genericStopInformerMap) Create(name string, informer stopinformer.Interface) {
	g.m.Lock()
	defer g.m.Unlock()
	g.informers[name] = informer
}

// Starts a new named stop informer and returns it for method chaining
func (g *genericStopInformerMap) CreateAndReturn(name string, informer stopinformer.Interface) stopinformer.Interface {
	g.m.Lock()
	defer g.m.Unlock()
	g.informers[name] = informer
	return informer
}

// Returns a stop informer, if a stop informer cannot be found, we return a new
// stop informer that resolves immediately in the event that the caller is method
// chaining
func (g *genericStopInformerMap) Get(name string) stopinformer.Interface {
	g.m.RLock()
	defer g.m.RUnlock()
	if i, ok := g.informers[name]; ok {
		return i
	}
	return stopinformer.NewGenericStopInformer().ResolveImmediately()
}

// Returns whether a stop informer exists
func (g *genericStopInformerMap) Exists(name string) bool {
	g.m.RLock()
	defer g.m.RUnlock()
	if _, ok := g.informers[name]; ok {
		return true
	}
	return false
}

// Iterates over all stop informers and sends the stop notification to each; this method
// blocks until all informers acknowledge the stop command
func (g *genericStopInformerMap) StopAll() {
	g.m.Lock()
	defer g.m.Unlock()
	for key, informer := range g.informers {
		informer.Stop()
		delete(g.informers, key)
	}
}

// Iterates over all stop informers and sends the stop notification to each; this method
// returns a channel and sends a value when all informers acknowledge they're stopped
func (g *genericStopInformerMap) StopAllAndNotify(buffer int) stopinformer.SingleStructChan {
	g.m.Lock()
	defer g.m.Unlock()

	notifyChan := make(stopinformer.SingleStructChan, buffer)

	wg := &sync.WaitGroup{}
	for key, informer := range g.informers {
		wg.Add(1)
		key := key
		go func(informer stopinformer.Interface) {
			defer wg.Done()
			defer func(key string) {
				g.m.Lock()
				defer g.m.Unlock()
				delete(g.informers, key)
			}(key)
			<-informer.StopAndNotify(0)
		}(informer)
	}

	go func() {
		wg.Wait()
		notifyChan <- struct{}{}
	}()

	return notifyChan
}
