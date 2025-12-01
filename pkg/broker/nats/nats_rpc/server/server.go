package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"library-service/pkg/broker/nats/nats_rpc"

	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"

	"time"
)

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 2 * time.Second
)

type CallHandler func(*nats.Msg) (interface{}, error)

type Server struct {
	ctx          context.Context
	eg           *errgroup.Group
	subject      string
	connection   *nats.Conn
	subscription *nats.Subscription
	router       map[string]CallHandler
	stop         chan struct{}
	notify       chan error
	timeout      time.Duration
}

func New(
	url,
	serverSubject string,
	router map[string]CallHandler,
	opts ...Option,
) (*Server, error) {
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(1) // Run only one goroutine

	connection, err := nats.Connect(
		url,
		nats.ReconnectWait(_defaultWaitTime),
		nats.MaxReconnects(_defaultAttempts),
		nats.Timeout(_defaultWaitTime),
	)
	if err != nil {
		return nil, fmt.Errorf("nats_rpc server - NewServer - nats.Connect: %w", err)
	}

	s := &Server{
		ctx:        ctx,
		eg:         group,
		subject:    serverSubject,
		connection: connection,
		router:     router,
		stop:       make(chan struct{}),
		notify:     make(chan error, 1),
		timeout:    _defaultTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Start -.
func (s *Server) Start() {
	s.eg.Go(func() error {
		err := s.subscribe()
		if err != nil {
			s.notify <- err

			close(s.notify)

			return err
		}

		// Wait for stop signal
		<-s.stop

		return nil
	})

	fmt.Println("nats_rpc server - Server - Started")
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	var shutdownErrors []error

	close(s.stop)

	err := s.eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {

		shutdownErrors = append(shutdownErrors, err)
	}

	// Unsubscribe
	if s.subscription != nil {
		err := s.subscription.Unsubscribe()
		if err != nil {

			shutdownErrors = append(shutdownErrors, err)
		}
	}

	// Close connection
	s.connection.Close()

	return errors.Join(shutdownErrors...)
}

func (s *Server) subscribe() error {
	subscription, err := s.connection.Subscribe(s.subject, s.handleMessage)
	if err != nil {
		return fmt.Errorf("nats_rpc server - subscribe - s.conn.AttemptConnect: %w", err)
	}

	s.subscription = subscription

	return nil
}

func (s *Server) handleMessage(msg *nats.Msg) {
	handler := msg.Header.Get("Handler")

	callHandler, ok := s.router[handler]
	if !ok {
		s.publish(msg, nil, nats_rpc.ErrBadHandler.Error())

		return
	}

	response, err := callHandler(msg)
	if err != nil {
		s.publish(msg, nil, nats_rpc.ErrInternalServer.Error())

		fmt.Errorf("nats_rpc server - Server - handleMessage - callHandler")
		return
	}

	body, err := json.Marshal(response)
	if err != nil {
		fmt.Println("nats_rpc server - Server - handleMessage - json.Marshal - callHandler")

		s.publish(msg, nil, nats_rpc.ErrInternalServer.Error())

		return
	}

	s.publish(msg, body, nats_rpc.Success)
}

func (s *Server) publish(msg *nats.Msg, body []byte, status string) {
	respondMsg := nats.NewMsg(msg.Reply)
	respondMsg.Header.Set("Status", status)
	respondMsg.Data = body

	err := s.connection.PublishMsg(respondMsg)
	if err != nil {
		fmt.Errorf("nats_rpc server - Server - publish - s.connection.PublishMsg: %w", err)
	}
}
