package stopinformers

import (
	"github.com/clarkmcc/stopinformer/stopinformer"
	"sync"
)

// StopInformers provides an easy interface for interacting with a set of stop informers.
// This method requires you to interact with all stop informers as a whole, if you need to
// interact individually with a set of stop informers, consider using StopInformerMap
type StopInformers interface {
	// Starts a new named stop informer
	Create(informer stopinformer.Interface)

	// Starts a new named stop informer and returns it for method chaining
	CreateAndReturn(informer stopinformer.Interface) stopinformer.Interface

	// Iterates over all stop informers and sends the stop notification to each; this method
	// blocks until all informers acknowledge the stop command
	StopAll()

	// Iterates over all stop informers and sends the stop notification to each; this method
	// returns a channel and sends a value when all informers acknowledge they're stopped
	StopAllAndNotify(buffer int) stopinformer.SingleStructChan
}

type genericStopInformers struct {
	informers []stopinformer.Interface
	m         *sync.RWMutex
}

func NewGenericStopInformers() StopInformers {
	return &genericStopInformers{
		informers: []stopinformer.Interface{},
		m:         &sync.RWMutex{},
	}
}

// Starts a new named stop informer
func (g *genericStopInformers) Create(informer stopinformer.Interface) {
	g.m.Lock()
	defer g.m.Unlock()
	g.informers = append(g.informers, informer)
}

// Starts a new named stop informer and returns it for method chaining
func (g *genericStopInformers) CreateAndReturn(informer stopinformer.Interface) stopinformer.Interface {
	g.m.Lock()
	defer g.m.Unlock()
	g.informers = append(g.informers, informer)
	return informer
}

// Iterates over all stop informers and sends the stop notification to each; this method
// blocks until all informers acknowledge the stop command
func (g *genericStopInformers) StopAll() {
	g.m.Lock()
	defer g.m.Unlock()
	for i, informer := range g.informers {
		informer.Stop()
		g.removeInformer(i)
	}
}

// Removes an informer from the genericStopInformers slice
func (g *genericStopInformers) removeInformer(index int) {
	g.informers[len(g.informers)-1],
		g.informers[index] = g.informers[index],
		g.informers[len(g.informers)-1]
}

// Iterates over all stop informers and sends the stop notification to each; this method
// returns a channel and sends a value when all informers acknowledge they're stopped
func (g *genericStopInformers) StopAllAndNotify(buffer int) stopinformer.SingleStructChan {
	g.m.Lock()
	defer g.m.Unlock()

	notifyChan := make(stopinformer.SingleStructChan, buffer)

	wg := &sync.WaitGroup{}
	for i, informer := range g.informers {
		wg.Add(1)
		i := i
		go func(informer stopinformer.Interface) {
			defer wg.Done()
			defer func(index int) {
				g.m.Lock()
				defer g.m.Unlock()
				g.removeInformer(i)
			}(i)
			<-informer.StopAndNotify(0)
		}(informer)
	}

	go func() {
		wg.Wait()
		notifyChan <- struct{}{}
	}()

	return notifyChan
}
