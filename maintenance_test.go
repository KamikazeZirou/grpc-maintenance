package grpc_maintenance

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"testing"
)

type AlwaysAvailableServer struct{}

func (a AlwaysAvailableServer) MaintenanceFuncOverride(fullMethodName string) bool {
	return false
}

type AlwaysUnavailableServer struct{}

func (a AlwaysUnavailableServer) MaintenanceFuncOverride(fullMethodName string) bool {
	return true
}

func TestUnaryServerInterceptor(t *testing.T) {
	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "mock", nil
	}

	type args struct {
		opts []Option
		info *grpc.UnaryServerInfo
	}

	type wants struct {
		err error
	}

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "By default, none of the gRPC services are considered to be under maintenance",
			args: args{
				opts: nil,
				info: &grpc.UnaryServerInfo{
					Server: "notOverrideServer",
				},
			},
			wants: wants{
				err: nil,
			},
		},
		{
			name: "If all services are under maintenance, an Unavailable error should be returned",
			args: args{
				opts: []Option{
					WithAlwaysMaintenance(),
				},
				info: &grpc.UnaryServerInfo{
					Server: "notOverrideServer",
				},
			},
			wants: wants{
				err: status.Error(codes.Unavailable, defaultMessage),
			},
		},
		{
			name: "Change the message during maintenance",
			args: args{
				opts: []Option{
					WithAlwaysMaintenance(),
					WithMessage("foobar"),
				},
				info: &grpc.UnaryServerInfo{
					Server: "notOverrideServer",
				},
			},
			wants: wants{
				err: status.Error(codes.Unavailable, "foobar"),
			},
		},
		{
			name: "Make a Server available",
			args: args{
				opts: []Option{
					WithAlwaysMaintenance(),
				},
				info: &grpc.UnaryServerInfo{
					Server: AlwaysAvailableServer{},
				},
			},
			wants: wants{
				err: nil,
			},
		},
		{
			name: "Put a Server under maintenance",
			args: args{
				opts: nil,
				info: &grpc.UnaryServerInfo{
					Server: AlwaysUnavailableServer{},
				},
			},
			wants: wants{
				err: status.Error(codes.Unavailable, defaultMessage),
			},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()
		t.Run(tt.name, func(t *testing.T) {
			interceptor := UnaryServerInterceptor(tt.args.opts...)
			res, err := interceptor(ctx, "mock", tt.args.info, mockHandler)

			if !reflect.DeepEqual(err, tt.wants.err) {
				t.Errorf("err %v, want %v", err, tt.wants.err)
			}

			if tt.wants.err != nil && res != nil {
				t.Errorf("res isn't nil although err is nil")
			}
		})
	}
}
