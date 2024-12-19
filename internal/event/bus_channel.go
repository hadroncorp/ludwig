package event

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/lo"

	"github.com/hadroncorp/ludwig/internal/tasking"
)

// BusChannel is a concrete implementation of the [Bus] interface.
//
// This component uses Go stdlib unbuffered channels to multiplex [Event](s).
//
// NOTE: Do not forget to call the Shutdown routine to gracefully finish all in-flight child processes.
type BusChannel struct {
	config                        BusConfig
	topicChannels                 map[string]chan ctxEventPair
	topicChannelsLock             sync.RWMutex
	inFlightBusProcWaitGroup      sync.WaitGroup
	isClosed                      atomic.Bool
	topicListenerWorkerNum        map[string]int
	inFlightListenerProcWaitGroup sync.WaitGroup
}

type ctxEventPair struct {
	ctx   context.Context
	event Event
}

const (
	_listenerExecTimeout = time.Second * 30
)

var (
	// compile-time interface assertions
	_ Bus                = (*BusChannel)(nil)
	_ tasking.Shutdowner = (*BusChannel)(nil)
)

// NewBusChannel allocates a new [BusChannel] instance.
func NewBusChannel(opts ...BusOption) *BusChannel {
	config := BusConfig{}
	for _, opt := range opts {
		opt(&config)
	}
	return &BusChannel{
		config:                        config,
		topicChannels:                 make(map[string]chan ctxEventPair),
		topicChannelsLock:             sync.RWMutex{},
		inFlightBusProcWaitGroup:      sync.WaitGroup{},
		topicListenerWorkerNum:        make(map[string]int),
		inFlightListenerProcWaitGroup: sync.WaitGroup{},
		isClosed:                      atomic.Bool{},
	}
}

// Shutdown stops in-flight processes gracefully.
func (b *BusChannel) Shutdown(ctx context.Context) error {
	if b.isClosed.Load() {
		return ErrBusClosed
	}
	b.isClosed.Store(true)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	b.inFlightBusProcWaitGroup.Wait()
	b.inFlightListenerProcWaitGroup.Wait()
	for topic, ch := range b.topicChannels {
		close(ch)
		delete(b.topicChannels, topic)
	}
	return nil
}

// Publish propagates `event` to a set of subscribers. If no subscribers are registered to the [Event.Topic],
// then an ErrNoSubscribers error will be returned.
func (b *BusChannel) Publish(ctx context.Context, event Event) error {
	if b.isClosed.Load() {
		return ErrBusClosed
	}
	b.inFlightBusProcWaitGroup.Add(1)
	defer b.inFlightBusProcWaitGroup.Done()
	b.topicChannelsLock.RLock()
	topicChannel, ok := b.topicChannels[event.Topic()]
	if !ok {
		b.topicChannelsLock.RUnlock()
		return ErrNoSubscribers
	}
	// this latch (wait group) will ensure all listener workers of the current event topic
	// get executed.
	b.inFlightListenerProcWaitGroup.Add(b.topicListenerWorkerNum[event.Topic()])
	b.topicChannelsLock.RUnlock()
	topicChannel <- ctxEventPair{
		ctx:   ctx,
		event: event,
	}
	return nil
}

// Subscribe registers a [ListenerFunc] routine to the given `topic`. The given listener routine
// will be executed everytime a new [Event] is published.
//
// Moreover, this routine has accepts a set of [ListenerInterceptorFunc] through a variadic argument.
// These interceptors will be executed after the global set of interceptors (defined by
// [BusConfig.ListenerInterceptors]).
//
// Finally, take into consideration that [Bus] processes might be concurrent; therefore,
// the user MUST call Publish routine after a Subscribe routine call to ensure this subscription gets properly executed.
func (b *BusChannel) Subscribe(_ context.Context, topic string, listenerFunc ListenerFunc, interceptors ...ListenerInterceptorFunc) error {
	if b.isClosed.Load() {
		return ErrBusClosed
	}
	b.inFlightBusProcWaitGroup.Add(1)
	defer b.inFlightBusProcWaitGroup.Done()
	b.topicChannelsLock.RLock()
	topicChannel, ok := b.topicChannels[topic]
	if !ok {
		b.topicChannelsLock.RUnlock()
		b.topicChannelsLock.Lock()
		topicChannel = make(chan ctxEventPair)
		b.topicChannels[topic] = topicChannel
		b.topicListenerWorkerNum[topic] = 1
		b.topicChannelsLock.Unlock()
	} else {
		b.topicChannelsLock.RUnlock()
		b.topicChannelsLock.Lock()
		b.topicListenerWorkerNum[topic] = b.topicListenerWorkerNum[topic] + 1
		b.topicChannelsLock.Unlock()
	}

	// enforces global interceptors to get executed first
	for _, interceptor := range interceptors {
		listenerFunc = interceptor(listenerFunc)
	}
	for _, interceptor := range b.config.ListenerInterceptors {
		listenerFunc = interceptor(listenerFunc)
	}
	go func() {
		for pair := range topicChannel {
			b.execListener(pair, listenerFunc)
		}
	}()
	return nil
}

func (b *BusChannel) execListener(pair ctxEventPair, listenerFunc ListenerFunc) {
	defer func() {
		// ensure bus gracefully handles panics by freeing in-flight listener proc latches to avoid deadlocks.
		if err := recover(); err != nil {
			b.inFlightListenerProcWaitGroup.Done()
			return
		}
		b.inFlightListenerProcWaitGroup.Done()
	}()
	ctx, cancelFunc := context.WithTimeout(pair.ctx, lo.CoalesceOrEmpty(b.config.ListenerTimeout, _listenerExecTimeout))
	defer cancelFunc()
	_ = listenerFunc(ctx, pair.event)
}
