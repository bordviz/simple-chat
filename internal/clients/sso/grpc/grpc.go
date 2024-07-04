package grpc

import (
	"log/slog"
	"simple-chat/internal/config"
	"simple-chat/internal/lib/logger/sl"

	ssov1 "github.com/bordviz/sso-protos/gen/go/sso"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Api ssov1.AuthClient
}

func NewClient(
	log *slog.Logger,
	cfg config.SSOClient,
) (*Client, error) {
	const op = "client.sso.grpc.NewClient"

	retryOpts := []grpcRetry.CallOption{
		grpcRetry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcRetry.WithMax(uint(cfg.RetriesCount)),
		grpcRetry.WithPerRetryTimeout(cfg.Timeout),
	}

	cc, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcRetry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		log.Error("failed to create grpc client", sl.OpErr(op, err))
		return nil, err
	}

	return &Client{
		Api: ssov1.NewAuthClient(cc),
	}, nil
}
