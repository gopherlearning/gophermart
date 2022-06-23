package rpc

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gopherlearning/gophermart/internal/repository"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type publicServer struct {
	v1.UnimplementedPublicServer
	db    repository.Storage
	loger logrus.FieldLogger
}

// NewPublicServer возвращает сервер
func NewPublicServer(db repository.Storage, loger logrus.FieldLogger) v1.PublicServer {
	return &publicServer{db: db, loger: loger}
}

func (s *publicServer) UsersRegister(ctx context.Context, req *v1.AuthRequest) (*v1.Empty, error) {
	claim, err := s.db.CreateUser(ctx, req.Login, req.Password)
	if err != nil {
		return nil, status.Error(http.StatusConflict, err.Error())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodPS256, claim)
	tokenString, err := token.SignedString(s.db.SigningKey())
	if err != nil {
		return nil, status.Error(http.StatusConflict, err.Error())
	}
	cookie := &http.Cookie{Value: tokenString, Name: "accesstoken", Path: "/", Expires: claim.ExpiresAt.Time}
	header := metadata.Pairs("Set-Cookie", cookie.String())
	err = grpc.SetHeader(ctx, header)
	if err != nil {
		return nil, status.Error(http.StatusConflict, err.Error())
	}
	return &v1.Empty{}, nil
}

func (s *publicServer) UsersLogin(ctx context.Context, req *v1.AuthRequest) (*v1.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
