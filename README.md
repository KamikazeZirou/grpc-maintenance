# grpc-maintenance

This module provides a grpc interceptor that turns on/off the maintenance state of grpc-server.

## Usage

### Put the entire grpc-server under maintenance

```go
grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        grpc_maintenance.UnaryServerInterceptor(grpc_maintenance.WithAlwaysMaintenance()),
    ),
)
```

This interceptor returns Unavailable error for all requests.

### Make only some APIs available

The configuration of the grpc server is the same as before.

```go
grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        grpc_maintenance.UnaryServerInterceptor(grpc_maintenance.WithAlwaysMaintenance()),
    ),
)
```

Then, implement MaintenanceFuncOverride in the handler of API.

```go
type HealthHandler struct{}

func NewHealthHandler() grpc_health_v1.HealthServer {
	return &HealthHandler{}
}

func (h *HealthHandler) MaintenanceFuncOverride(fullMethodName string) bool {
	return false
}

func (h *HealthHandler) Check(context.Context, *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
    return &grpc_health_v1.HealthCheckResponse{
    Status: grpc_health_v1.HealthCheckResponse_SERVING,
    }, nil
}

...
```

In this example, only HealthHandler.Check() is available.