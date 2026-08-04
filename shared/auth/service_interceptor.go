package auth

import (
	"context"
	"crypto/subtle"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	serviceTokenHeader = "x-service-token"
	serviceNameHeader  = "x-service-name"
)

type ServiceAuthInterceptor struct {
	expectedToken   string
	allowedServices map[string]struct{}
}

func NewServiceAuthInterceptor(
	expectedToken string,
	allowedServices ...string,
) *ServiceAuthInterceptor {
	services := make(map[string]struct{}, len(allowedServices))

	for _, service := range allowedServices {
		services[service] = struct{}{}
	}

	return &ServiceAuthInterceptor{
		expectedToken:   expectedToken,
		allowedServices: services,
	}
}

type ServiceCredentialsInterceptor struct {
	serviceName  string
	serviceToken string
}

func NewServiceCredentialsInterceptor(
	serviceName string,
	serviceToken string,
) *ServiceCredentialsInterceptor {
	return &ServiceCredentialsInterceptor{
		serviceName:  serviceName,
		serviceToken: serviceToken,
	}
}

func (i *ServiceAuthInterceptor) authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(
			codes.Unauthenticated,
			"missing req metadata")
	}
	servicesNames := md.Get(serviceNameHeader)
	serviceTokens := md.Get(serviceTokenHeader)

	if len(servicesNames) == 0 || len(serviceTokens) == 0 {
		return status.Error(
			codes.Unauthenticated,
			"missing service credentials",
		)
	}

	serviceName := servicesNames[0]
	serviceToken := serviceTokens[0]

	if _, allowed := i.allowedServices[serviceName]; !allowed {
		return status.Error(
			codes.PermissionDenied,
			"service is not allowed",
		)
	}

	if subtle.ConstantTimeCompare(
		[]byte(serviceToken),
		[]byte(i.expectedToken),
	) != 1 {
		return status.Error(
			codes.Unauthenticated,
			"invalid service token",
		)
	}

	return nil

}

func (i *ServiceAuthInterceptor) Unary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if err := i.authenticate(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)

}

func (i *ServiceAuthInterceptor) Stream(
	srv any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if err := i.authenticate(stream.Context()); err != nil {
		return err
	}

	return handler(srv, stream)
}

func (i *ServiceCredentialsInterceptor) Unary(
	ctx context.Context,
	method string,
	req any,
	reply any,
	connection *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		serviceNameHeader,
		i.serviceName,
		serviceTokenHeader,
		i.serviceToken,
	)

	return invoker(
		ctx,
		method,
		req,
		reply,
		connection,
		opts...,
	)
}

func (i *ServiceCredentialsInterceptor) Stream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	connection *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		serviceNameHeader,
		i.serviceName,
		serviceTokenHeader,
		i.serviceToken,
	)

	return streamer(
		ctx,
		desc,
		connection,
		method,
		opts...,
	)
}
