package cmd

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func newGRPCConnection(ctx context.Context, addr string, insecure bool) (*grpc.ClientConn, error) {
	var dialOpts []grpc.DialOption
	if insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	return grpc.DialContext(ctx, addr, dialOpts...)
}
