package cmd

import (
	"context"
	"time"

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

	dialOpts = append(dialOpts, grpc.WithBlock(), grpc.WithBackoffMaxDelay(time.Second))
	return grpc.DialContext(ctx, addr, dialOpts...)
}
