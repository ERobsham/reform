package streams

import (
	"context"
	"sync"
)

func NewStreamAggregator(ctx context.Context, streams []InputStream) *StreamAggregator {
	newCtx, cancelFn := context.WithCancel(ctx)

	a := &StreamAggregator{
		ctx:        newCtx,
		cancelFunc: cancelFn,
		wg:         sync.WaitGroup{},
		output:     make(chan string),
	}

	a.wg.Add(len(streams))
	for _, stream := range streams {
		go a.pullInput(newCtx, stream)
	}

	go a.awaitDone()

	return a
}

type StreamAggregator struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup

	output chan string
}

func (a *StreamAggregator) Next() (string, error) {
	str, ok := <-a.output
	if !ok {
		return str, ErrStreamClosed
	} else {
		return str, nil
	}
}

func (a *StreamAggregator) Close() {
	a.cancelFunc()
	a.wg.Wait()
}

func (a *StreamAggregator) awaitDone() {
	a.wg.Wait()
	close(a.output)
}

func (a *StreamAggregator) pullInput(ctx context.Context, stream InputStream) {
	defer a.wg.Done()
	for {
		val, err := stream.Next()
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case a.output <- val:
		}
	}
}
