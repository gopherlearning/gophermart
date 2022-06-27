package rpc

import (
	"context"
	"net/http"

	"github.com/gopherlearning/gophermart/internal/repository"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type privateServer struct {
	v1.UnimplementedPrivateServer
	db    repository.Storage
	loger logrus.FieldLogger
}

// NewPublicServer возвращает сервер, не требующий авторизации
func NewPrivateServer(db repository.Storage, loger logrus.FieldLogger) v1.PrivateServer {
	return &privateServer{db: db, loger: loger}
}
func (s *privateServer) OrdersAdd(ctx context.Context, req *v1.OrderRequest) (*v1.Empty, error) {
	err := s.db.CreateOrder(ctx, req.Order)
	if err != nil {
		return nil, err
	}
	return &v1.Empty{}, status.Error(codes.OK, "")
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
