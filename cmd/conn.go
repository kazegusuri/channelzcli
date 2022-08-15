package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TLSData struct {
	CAPool     string
	ClientCert string
	ClientKey  string
}

func newGRPCConnection(ctx context.Context, addr string, insecure bool, tlsData TLSData) (*grpc.ClientConn, error) {
	var dialOpts []grpc.DialOption
	if insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		var ca *x509.CertPool
		var err error
		if tlsData.CAPool == "" {
			ca, err = x509.SystemCertPool()
			if err != nil {
				return nil, fmt.Errorf("can't load system cert pool: %v", err)
			}
		} else {
			c, err := os.ReadFile(tlsData.CAPool)
			if err != nil {
				return nil, fmt.Errorf("could not read CA from %q: %v", tlsData.CAPool, err)
			}
			ca = x509.NewCertPool()
			if !ca.AppendCertsFromPEM(c) {
				return nil, fmt.Errorf("could not add CA cert to pool: %v", err)
			}
		}
		cert, err := tls.LoadX509KeyPair(tlsData.ClientCert, tlsData.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("can't load client cert: %v", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      ca,
			MinVersion:   tls.VersionTLS13,
		})))
	}

	dialOpts = append(dialOpts, grpc.WithBlock(), grpc.WithBackoffMaxDelay(time.Second))
	return grpc.DialContext(ctx, addr, dialOpts...)
}
