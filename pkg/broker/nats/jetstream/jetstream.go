package jetstream

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	_defaultTimeout = 5 * time.Second
)

type JetStream struct {
	nc *nats.Conn
	js jetstream.JetStream
}

type Config struct {
	URL           string
	StreamName    string
	Subjects      []string
	MaxAge        time.Duration
	MaxBytes      int64
	Replicas      int
	StorageType   jetstream.StorageType
	RetentionType jetstream.RetentionPolicy
}

func New(cfg Config) (*JetStream, error) {
	nc, err := nats.Connect(
		cfg.URL,
		nats.ReconnectWait(5*time.Second),
		nats.MaxReconnects(10),
	)
	if err != nil {
		return nil, fmt.Errorf("jetstream - New - nats.Connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream - New - jetstream.New: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), _defaultTimeout)
	defer cancel()

	streamConfig := jetstream.StreamConfig{
		Name:        cfg.StreamName,
		Subjects:    cfg.Subjects,
		MaxAge:      cfg.MaxAge,
		MaxBytes:    cfg.MaxBytes,
		Storage:     cfg.StorageType,
		Retention:   cfg.RetentionType,
		Replicas:    cfg.Replicas,
		Compression: jetstream.S2Compression,
	}

	_, err = js.CreateStream(ctx, streamConfig)
	if err != nil {
		_, err = js.UpdateStream(ctx, streamConfig)
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("jetstream - New - CreateStream: %w", err)
		}
	}

	return &JetStream{
		nc: nc,
		js: js,
	}, nil
}

func (j *JetStream) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := j.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("jetstream - Publish: %w", err)
	}
	return nil
}

func (j *JetStream) PublishAsync(subject string, data []byte) (jetstream.PubAckFuture, error) {
	future, err := j.js.PublishAsync(subject, data)
	if err != nil {
		return nil, fmt.Errorf("jetstream - PublishAsync: %w", err)
	}
	return future, nil
}

func (j *JetStream) CreateConsumer(ctx context.Context, streamName, consumerName string, subjects []string) (jetstream.Consumer, error) {
	consumerConfig := jetstream.ConsumerConfig{
		Name:          consumerName,
		Durable:       consumerName,
		FilterSubjects: subjects,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    3,
		AckWait:       30 * time.Second,
	}

	consumer, err := j.js.CreateOrUpdateConsumer(ctx, streamName, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("jetstream - CreateConsumer: %w", err)
	}

	return consumer, nil
}

func (j *JetStream) ConsumeMessages(ctx context.Context, consumer jetstream.Consumer, handler func(jetstream.Msg) error) error {
	_, err := consumer.Consume(func(msg jetstream.Msg) {
		if err := handler(msg); err != nil {
			msg.Nak()
			return
		}
		msg.Ack()
	})

	if err != nil {
		return fmt.Errorf("jetstream - ConsumeMessages: %w", err)
	}

	<-ctx.Done()
	return nil
}

func (j *JetStream) Close() {
	if j.nc != nil {
		j.nc.Close()
	}
}
