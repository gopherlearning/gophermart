package repository

import (
	"context"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"

	v1 "github.com/gopherlearning/gophermart/proto/v1"
)

type Claim struct {
	jwt.RegisteredClaims
}

type Storage interface {
	CreateUser(ctx context.Context, login, password string) (*Claim, error)
	GetUser(ctx context.Context, login, password string) (*Claim, error)
	GetUserBySession(context.Context, *Claim) (*Claim, error)
	CreateOrder(ctx context.Context, id int64) error
	GetBalance(ctx context.Context) (*v1.Balance, error)
	GetOrders(ctx context.Context) ([]*v1.Order, error)
	CreateWithdraw(ctx context.Context, id string, sum float64) error
	GetWithdrawals(ctx context.Context) ([]*v1.WithdrawRequest, error)
	UnaryCheckToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
	StreamCheckToken(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
	SigningKey() string
	AccrualMonitor(context.Context, *sync.WaitGroup, string)
}
