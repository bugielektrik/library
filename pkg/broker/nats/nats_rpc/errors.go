package nats_rpc

import "errors"

var (
	ErrTimeout        = errors.New("timeout")
	ErrInternalServer = errors.New("internal server error")
	ErrBadHandler     = errors.New("bad handler")
)

const Success = "success"
