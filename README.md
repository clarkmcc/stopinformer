# Stop Informer
Stop informer provides a simple and concise way to maintain an exit-handle on a running goroutine. It also allows you to wait until the stop is acknowledged by the goroutine using either a blocking stop, or using a notification channel. When paired with stopinformermap, you can maintain named handles on multiple goroutines, stop all of them.

## Installation
```bash
$ go get github.com/clarkmcc/stopinformer
```

## Example
### Single Stop Informer
For managing a single stop informer.
```go
func main() {

    // Create a new informer
    informer := NewGenericStopInformer()

    // Start a goroutine and perform a non blocking watch operation
    // on the informer.Watch() channel
    go func() {
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
    }()
    
    // Perform a blocking stop of the goroutine
    informer.Stop()

    // Or get a notification via a channel when the goroutine has acknowledged the stop
    <- informer.StopAndNotify(0)
}
```

### Multiple Stop Informers
For managing multiple (non-named) stop informers. This is useful for when you want to interact with multipl stop informers as a group.

```go
func main() {
    informers := NewGenericStopInformers()
    
    informers.Create(stopinformer.NewGenericStopInformer())
    informers.Create(stopinformer.NewGenericStopInformer())

    // Pass your informer into goroutines here...

    // Block until all goroutines acknowledge the stop
    informers.StopAll()

    //Or get a notification via a channel when all goroutines acknowledged the stop
    <-informers.StopAllAndNotify(0)
}
```

### Stop Informer Map
For managing multiple named stop informers. This is useful for when you want to interact with multiple stop informers individually.
```go
func main() {
    informers := NewGenericStopInformerMap()
    
    informers.Create("informer1", stopinformer.NewGenericStopInformer())
    informers.Create("informer2", stopinformer.NewGenericStopInformer())

    // Pass your informer into goroutines here...

    // Stop a single goroutine by name
    informers.Get("informer1").Stop()

    // Informers that don't exist resolve immediately to prevent existence checking and error handling
    informers.Get("doesntexist").Stop()

    // Or you can check for existence
    ok := informers.Exists("informer1")

    // Block until all goroutines acknowledge the stop
    informers.StopAll()

    //Or get a notification via a channel when all goroutines acknowledged the stop
    <-informers.StopAllAndNotify(0)
}
```
