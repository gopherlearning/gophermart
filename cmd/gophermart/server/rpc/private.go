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
		case errors.Is(err, repository.ErrWrongOrderNumber):
			return &v1.Empty{}, status.Error(codes.Code(http.StatusUnprocessableEntity), repository.ErrWrongOrderNumber.Error())
		}
		return nil, err
	}
	return &v1.Empty{}, status.Error(codes.Code(http.StatusAccepted), repository.ErrSuccessOrderUploaded.Error())
}

func (s *privateServer) OrdersGet(ctx context.Context, req *v1.Empty) (*v1.OrdersResponse, error) {
	orders, err := s.db.GetOrders(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, repository.ErrInternalServer.Error())
	}
	if len(orders) == 0 {
		return nil, status.Error(codes.Code(http.StatusNoContent), repository.ErrNoContent.Error())
	}
	return &v1.OrdersResponse{Orders: orders}, status.Error(codes.OK, "")
}

func (s *privateServer) GetBalance(ctx context.Context, _ *v1.Empty) (*v1.Balance, error) {
	balance, err := s.db.GetBalance(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, repository.ErrInternalServer.Error())
	}
	return balance, status.Error(codes.OK, "")
	// return nil, nil
}

func (s *privateServer) BalanceWithdraw(ctx context.Context, req *v1.WithdrawRequest) (*v1.Empty, error) {
	err := s.db.CreateWithdraw(ctx, req.Order, req.Sum)
	switch {
	case errors.Is(err, repository.ErrWrongOrderNumber):
		return nil, status.Error(codes.Code(http.StatusUnprocessableEntity), repository.ErrWrongOrderNumber.Error())
	case errors.Is(err, repository.ErrLowBalance):
		return nil, status.Error(codes.Code(http.StatusPaymentRequired), repository.ErrLowBalance.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, repository.ErrInternalServer.Error())
	}
	return &v1.Empty{}, status.Error(codes.OK, "")
}

func (s *privateServer) GetWithdrawals(ctx context.Context, _ *v1.Empty) (*v1.WithdrawsResponse, error) {
	withdrawals, err := s.db.GetWithdrawals(ctx)
	switch {
	case len(withdrawals) == 0:
		return nil, status.Error(codes.Code(http.StatusNoContent), repository.ErrNoWithdrawals.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, repository.ErrInternalServer.Error())
	}
	return &v1.WithdrawsResponse{Withdraws: withdrawals}, status.Error(codes.OK, "")
}
