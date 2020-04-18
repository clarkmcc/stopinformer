package stopinformer

type SingleStructChan chan struct{}
type SingleAckerChan chan *stopAcker
type DoubleStructChan chan chan struct{}

type Interface interface {
	// Stop blocks until the stop informer has received confirmation that the resource has stopped
	Stop()

	// StopAndNotify functions the same as Stop but instead of blocking, returns a chan that fires
	// when the requested resource has stopped
	StopAndNotify(buffer int) SingleStructChan

	// Watch is utilized by a goroutine--generally inside of a select statement--to received the
	// stop notification
	Watch() SingleAckerChan

	// Used by stopinformermap to wrap a stop informer with essentially a no-op behavior in the
	// event the the informer being looked up doesn't exist. This allows a caller to daisy chain
	// methods from the stopinformermap and stopinformer without worrying about handling errors
	ResolveImmediately() Interface
}

// genericStopInformer implements the Interface interface
type genericStopInformer struct {
	// Values are sent to this channel when the Stop or StopAndNotify methods are called, this
	// channel is used internally to coordinate the stop
	internalStopChan DoubleStructChan

	// This channel is usually sent as a value into the internalStopChan. Values are sent to this
	// channel when the goroutine being stopped has acknowledged the stop
	internalNotifyStoppedChan SingleStructChan

	// This channel is used by the goroutine to watch for stops. Values are are sent to this
	// channel when Stop and StopAndNotify are called
	internalWatchChan SingleAckerChan

	// This channel is used by callers to get stop acknowledgements from goroutines via a channel
	// as opposed to blocking on the Stop method. This channel is returned by StopAndNotify. This
	externalWatchStoppedChan SingleStructChan

	// Used by stopinformermap to wrap a stop informer with essentially a no-op behavior in the
	// event the the informer being looked up doesn't exist. This allows a caller to daisy chain
	// methods from the stopinformermap and stopinformer without worrying about handling errors
	resolveImmediately bool
}

// stopAcker is sent as a value through the internalStopChan and is generally received by the
// target goroutine
type stopAcker struct {
	ackChan SingleStructChan
}

// Returns a new instance of the stopAcker
func NewStopAcker(ackChan SingleStructChan) *stopAcker {
	return &stopAcker{ackChan: ackChan}
}

// Utilized by the target goroutine to 'acknowledge' the stop command. Calling this method
// unblocks the Interface.Stop command or sends a notification to the Interface.StopAndNotify
// channel. This is usually called using defer.
func (s *stopAcker) Acknowledge() {
	s.ackChan <- struct{}{}
}

// Creates a new stop informer with an optional parameter for the buffer size
func NewGenericStopInformer() Interface {
	internalStopChan := make(DoubleStructChan)
	internalNotifyStoppedChan := make(SingleStructChan)
	internalWatchChan := make(SingleAckerChan)
	externalWatchStoppedChan := make(SingleStructChan)

	go func() {
		<-internalStopChan
		internalWatchChan <- NewStopAcker(internalNotifyStoppedChan)
	}()

	return &genericStopInformer{
		resolveImmediately:        false,
		internalStopChan:          internalStopChan,
		internalNotifyStoppedChan: internalNotifyStoppedChan,
		internalWatchChan:         internalWatchChan,
		externalWatchStoppedChan:  externalWatchStoppedChan,
	}
}

// Stop blocks until the stop informer has received confirmation that the resource has stopped
func (g *genericStopInformer) Stop() {
	if !g.resolveImmediately {
		g.internalStopChan <- g.internalNotifyStoppedChan
		<-g.internalNotifyStoppedChan
	}
}

// Used by stopinformermap to wrap a stop informer with essentially a no-op behavior in the
// event the the informer being looked up doesn't exist. This allows a caller to daisy chain
// methods from the stopinformermap and stopinformer without worrying about handling errors
func (g *genericStopInformer) ResolveImmediately() Interface {
	g.resolveImmediately = true
	return g
}

// StopAndNotify functions the same as Stop but instead of blocking, returns a buffered channel
// that fires when the requested resource has stopped.
func (g *genericStopInformer) StopAndNotify(buffer int) SingleStructChan {
	g.externalWatchStoppedChan = make(SingleStructChan, buffer)
	if !g.resolveImmediately {
		go func() {
			g.internalStopChan <- g.internalNotifyStoppedChan
			<-g.internalNotifyStoppedChan
			g.externalWatchStoppedChan <- struct{}{}
		}()
	} else {
		g.externalWatchStoppedChan <- struct{}{}
	}
	return g.externalWatchStoppedChan
}

// Watch is utilized by a goroutine--generally inside of a select statement--to received the
// stop notification
func (g *genericStopInformer) Watch() SingleAckerChan {
	return g.internalWatchChan
}
