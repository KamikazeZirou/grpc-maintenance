package grpc_maintenance

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor create a unary server interceptors that returns Unavailable error if it is under maintenance.
//
// By default, none of the gRPC services are considered to be under maintenance.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := buildOptions(opts...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if overrideSrv, ok := info.Server.(MaintenanceFuncOverride); ok {
			if overrideSrv.MaintenanceFuncOverride(info.FullMethod) {
				return nil, status.Error(codes.Unavailable, o.message)
			}
		} else if o.maintenanceFunc() {
			return nil, status.Error(codes.Unavailable, o.message)
		}

		return handler(ctx, req)
	}
}

// MaintenanceFunc is the pluggable function that determines if maintenance is in progress.
type MaintenanceFunc func() bool

func alwaysAvailable() bool {
	return false
}

// MaintenanceFuncOverride allows a given gRPC service implementation to override the global `MaintenanceFunc`.
//
// If a service implements the MaintenanceFuncOverride method, it takes precedence over the `MaintenanceFunc` method,
// and will be called instead of MaintenanceFunc for all method invocations within that service.
type MaintenanceFuncOverride interface {
	MaintenanceFuncOverride(fullMethodName string) bool
}

type options struct {
	maintenanceFunc MaintenanceFunc
	message         string
}

const defaultMessage = "メンテナンス中です。しばらく待ってから再度アクセスをお願いします。"

type Option func(*options)

func buildOptions(opts ...Option) *options {
	o := &options{
		maintenanceFunc: alwaysAvailable,
		message:         defaultMessage,
	}

	for _, v := range opts {
		v(o)
	}

	return o
}

// WithAlwaysMaintenance sets all gRPC services to always be in maintenance.
func WithAlwaysMaintenance() Option {
	return func(o *options) {
		o.maintenanceFunc = func() bool {
			return true
		}
	}
}

// WithMessage sets the message when maintenance is in progress
func WithMessage(message string) Option {
	return func(o *options) {
		o.message = message
	}
}
