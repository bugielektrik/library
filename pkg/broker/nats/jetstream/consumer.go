package jetstream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type EventHandler func(Event) error

type Consumer struct {
	js       *JetStream
	logger   *zap.Logger
	handlers map[string]EventHandler
}

func NewConsumer(js *JetStream, logger *zap.Logger) *Consumer {
	return &Consumer{
		js:       js,
		logger:   logger,
		handlers: make(map[string]EventHandler),
	}
}

func (c *Consumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
	c.logger.Info("event handler registered", zap.String("event_type", eventType))
}

func (c *Consumer) Start(ctx context.Context, streamName, consumerName string, subjects []string) error {
	consumer, err := c.js.CreateConsumer(ctx, streamName, consumerName, subjects)
	if err != nil {
		return fmt.Errorf("consumer - Start - CreateConsumer: %w", err)
	}

	c.logger.Info("consumer started",
		zap.String("stream", streamName),
		zap.String("consumer", consumerName),
		zap.Strings("subjects", subjects),
	)

	return c.js.ConsumeMessages(ctx, consumer, c.handleMessage)
}

func (c *Consumer) handleMessage(msg jetstream.Msg) error {
	var event Event
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		c.logger.Error("failed to unmarshal event",
			zap.Error(err),
			zap.String("subject", msg.Subject()),
		)
		return err
	}

	c.logger.Debug("event received",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.Type),
		zap.String("subject", msg.Subject()),
	)

	handler, ok := c.handlers[event.Type]
	if !ok {
		c.logger.Warn("no handler for event type",
			zap.String("event_type", event.Type),
		)
		return nil
	}

	if err := handler(event); err != nil {
		c.logger.Error("handler failed",
			zap.Error(err),
			zap.String("event_type", event.Type),
			zap.String("event_id", event.ID),
		)
		return err
	}

	c.logger.Debug("event processed successfully",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.Type),
	)

	return nil
}
