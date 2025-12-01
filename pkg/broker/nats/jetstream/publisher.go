package jetstream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type Publisher struct {
	js     *JetStream
	logger *zap.Logger
	source string
}

func NewPublisher(js *JetStream, logger *zap.Logger, source string) *Publisher {
	return &Publisher{
		js:     js,
		logger: logger,
		source: source,
	}
}

func (p *Publisher) PublishEvent(ctx context.Context, subject, eventType string, data map[string]interface{}) error {
	event := Event{
		ID:        generateID(),
		Type:      eventType,
		Source:    p.source,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event",
			zap.Error(err),
			zap.String("event_type", eventType),
		)
		return fmt.Errorf("publisher - PublishEvent - json.Marshal: %w", err)
	}

	err = p.js.Publish(ctx, subject, eventData)
	if err != nil {
		p.logger.Error("failed to publish event",
			zap.Error(err),
			zap.String("subject", subject),
			zap.String("event_type", eventType),
		)
		return err
	}

	p.logger.Debug("event published",
		zap.String("subject", subject),
		zap.String("event_type", eventType),
		zap.String("event_id", event.ID),
	)

	return nil
}

func (p *Publisher) PublishBookCreated(ctx context.Context, bookID, title string) error {
	return p.PublishEvent(ctx, "events.book.created", "book.created", map[string]interface{}{
		"book_id": bookID,
		"title":   title,
	})
}

func (p *Publisher) PublishBookUpdated(ctx context.Context, bookID, title string) error {
	return p.PublishEvent(ctx, "events.book.updated", "book.updated", map[string]interface{}{
		"book_id": bookID,
		"title":   title,
	})
}

func (p *Publisher) PublishBookDeleted(ctx context.Context, bookID string) error {
	return p.PublishEvent(ctx, "events.book.deleted", "book.deleted", map[string]interface{}{
		"book_id": bookID,
	})
}

func (p *Publisher) PublishMemberCreated(ctx context.Context, memberID, fullName string) error {
	return p.PublishEvent(ctx, "events.member.created", "member.created", map[string]interface{}{
		"member_id": memberID,
		"full_name": fullName,
	})
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
