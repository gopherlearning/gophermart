package rpc

import (
	"context"
	"net/http"

	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type privateServer struct {
	v1.UnimplementedPrivateServer
	// db storage.Storage
	// subscriber Subscriber
}

// NewPublicServer возвращает сервер, не требующий авторизации
func NewPrivateServer() v1.PrivateServer {
	return &privateServer{}
}
func (s *privateServer) OrdersAdd(ctx context.Context, req *v1.OrderRequest) (*v1.Empty, error) {
	// logrus.Info(ctx)
	return nil, status.Error(codes.Unimplemented, "not implemented")
	// return nil, status.Error(codes.OK, "server/rpc/clients/All: count out of range")
}

func (s *privateServer) OrdersGet(ctx context.Context, req *v1.Empty) (*v1.OrdersResponse, error) {
	if 1 == 2 {
		return nil, status.Error(http.StatusNoContent, "")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(http.StatusForbidden, "")
	}
	logrus.Info(md)
	return nil, status.Error(codes.Unimplemented, "not implemented")
	// return &v1.OrdersResponse{Orders: []*v1.Order{{Number: "14212312", Status: v1.Order_PROCESSED}}}, nil
	// status.Error(http.StatusOK, "")
}

func (s *privateServer) GetBalance(ctx context.Context, _ *v1.Empty) (*v1.Balance, error) {
	logrus.Info(ctx)
	return nil, status.Error(codes.Unimplemented, "not implemented")
	// return nil, nil
}

func (s *privateServer) BalanceWithdraw(ctx context.Context, req *v1.WithdrawRequest) (*v1.Empty, error) {
	logrus.Info(ctx)
	return nil, status.Error(codes.Unimplemented, "not implemented")
	// return nil, nil
}

func (s *privateServer) GetWithdrawals(ctx context.Context, _ *v1.Empty) (*v1.WithdrawsResponse, error) {
	logrus.Info(ctx)
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
