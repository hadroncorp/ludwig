package event_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/hadroncorp/ludwig/internal/event"
)

func TestNewBusChannel(t *testing.T) {
	ctxRoot := context.Background()
	topic := "org.acme.users"
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var interceptor event.ListenerInterceptorFunc = func(next event.ListenerFunc) event.ListenerFunc {
		return func(ctx context.Context, ev event.Event) error {
			logger.InfoContext(ctx, "executing listener func",
				slog.String("topic", ev.Topic()), slog.String("msg", string(ev.Message())))
			err := next(ctx, ev)
			if err != nil {
				logger.ErrorContext(ctx, "listener error",
					slog.String("topic", ev.Topic()), slog.String("err", err.Error()))
				return err
			}
			return nil
		}
	}
	bus := event.NewBusChannel(event.WithListenerInterceptors(interceptor))
	defer func() {
		if err := bus.Shutdown(ctxRoot); err != nil {
			logger.Error("error at bus shutdown", slog.String("err", err.Error()))
		}
	}()
	var interceptorChild event.ListenerInterceptorFunc = func(next event.ListenerFunc) event.ListenerFunc {
		return func(ctx context.Context, ev event.Event) error {
			logger.InfoContext(ctx, "executing listener func, interceptor child",
				slog.String("topic", ev.Topic()), slog.String("msg", string(ev.Message())))
			val, ok := ctx.Value("some-key").(string)
			if ok {
				logger.InfoContext(ctx, "interceptor child got ctx value", slog.String("val", val))
			}
			return next(ctx, ev)
		}
	}
	_ = bus.Subscribe(ctxRoot, topic, func(ctx context.Context, ev event.Event) error {
		t.Logf("received message: %s", ev.Message())
		return errors.New("test error")
	}, interceptorChild)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		ctxPublish := context.WithValue(ctxRoot, "some-key", "some-value")
		ev := event.NewEvent(event.WithTopic(topic), event.WithMessage([]byte("test message")))
		err := bus.Publish(ctxPublish, ev)
		if err != nil {
			t.Error(err)
		}
	}()
	go func() {
		defer wg.Done()
		ev := event.NewEvent(event.WithTopic(topic), event.WithMessage([]byte("test message v2")))
		err := bus.Publish(ctxRoot, ev)
		if err != nil {
			t.Error(err)
		}
	}()
	wg.Wait()
}
