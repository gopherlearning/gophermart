package rpc

import (
	"context"
	"errors"
	"net/http"

	"github.com/gopherlearning/gophermart/internal/repository"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
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
		switch {
		case errors.Is(err, repository.ErrOrderAlreadyUploadedOther):
			return &v1.Empty{}, status.Error(codes.Aborted, repository.ErrOrderAlreadyUploadedOther.Error())
		case errors.Is(err, repository.ErrOrderAlreadyUploaded):
			return &v1.Empty{Message: repository.ErrOrderAlreadyUploaded.Error()}, status.Error(codes.OK, "")
		case errors.Is(err, repository.ErrWrongFormat):
			return &v1.Empty{}, status.Error(codes.InvalidArgument, repository.ErrWrongFormat.Error())
		}
		return nil, err
	}
	return &v1.Empty{}, status.Error(codes.Code(http.StatusAccepted), repository.ErrSuccessOrderUploaded.Error())
}

func (s *privateServer) OrdersGet(ctx context.Context, req *v1.Empty) (*v1.OrdersResponse, error) {
	orders, err := s.db.GetOrders(ctx)
	if err != nil {
		switch {
		case len(orders) == 0:
			return nil, status.Error(codes.Code(http.StatusNoContent), repository.ErrNoContent.Error())
		default:
			return nil, status.Error(codes.Internal, repository.ErrInternalServer.Error())
		}
	}
	return &v1.OrdersResponse{Orders: orders}, status.Error(codes.OK, "")
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
