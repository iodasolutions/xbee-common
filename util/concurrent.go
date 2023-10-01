package util

import (
	"context"
	"sync"
)

func Multiplex[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	multiplexedStream := make(chan T)

	multiplex := func(c <-chan T) {
		defer wg.Done()
		for i := range c {
			select {
			case <-ctx.Done():
				return
			case multiplexedStream <- i:
			}
		}
	}

	// Select from all the channels
	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	// Wait for all the reads to complete
	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}

type Executor func(ctx context.Context) *XbeeError

// Execute run functions concurrently, and returns any error encountered by these functions.
func Execute(ctx context.Context, funcs ...Executor) *XbeeError {
	if len(funcs) == 0 {
		return nil
	}
	var errors []*XbeeError
	errCh := make(chan *XbeeError)
	waitCh := make(chan bool)
	go func() {
		for {
			select {
			case err, ok := <-errCh:
				if ok {
					errors = append(errors, err)
				} else {
					waitCh <- true
					return
				}
			}
		}
	}()
	var wg sync.WaitGroup
	wg.Add(len(funcs))
	for _, f := range funcs {
		go func(f Executor) {
			defer wg.Done()
			if err := f(ctx); err != nil {
				errCh <- err
			}
		}(f)
	}
	wg.Wait()
	close(errCh)
	<-waitCh
	if len(errors) > 0 {
		return CauseBy(errors...)
	}
	return nil
}
