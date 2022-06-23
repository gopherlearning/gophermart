package repository

import (
	"context"

	"github.com/golang-jwt/jwt/v4"

	v1 "github.com/gopherlearning/gophermart/proto/v1"
)

type Claim struct {
	jwt.RegisteredClaims
}

type Storage interface {
	CreateUser(ctx context.Context, login, password string) (*Claim, error)
	GetUser(ctx context.Context, login, password string) (*Claim, error)
	GetUserBySession(context.Context, *Claim) (*Claim, error)
	CreateOrder(ctx context.Context, id string) error
	GetOrder(ctx context.Context, id string) (*v1.Order, error)
	GetOrders(ctx context.Context, id string) ([]*v1.Order, error)
	Withdrawn(ctx context.Context, id string, sum float64)
	CheckToken(ctx context.Context) (context.Context, error)
	SigningKey() interface{}
}
