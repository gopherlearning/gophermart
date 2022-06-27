package rpc

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gopherlearning/gophermart/internal/repository"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
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
		if strings.Contains(err.(*pgconn.PgError).Message, "duplicate key value violates unique constraint") {
			return nil, status.Error(codes.Aborted, repository.ErrLoginConflict.Error())
		}
		return nil, status.Error(codes.InvalidArgument, repository.ErrWrongFormat.Error())
	}
	e, err := s.setCookie(ctx, req, claim)
	if err != nil {
		return nil, status.Error(codes.Internal, repository.ErrInternalServer.Error())
	}
	return e, status.New(codes.OK, repository.ErrSuccessRegistered.Error()).Err()
}

func (s *publicServer) UsersLogin(ctx context.Context, req *v1.AuthRequest) (*v1.Empty, error) {
	if len(req.Login) < 4 || len(req.Password) < 6 {
		return nil, status.Error(codes.InvalidArgument, repository.ErrWrongFormat.Error())
	}
	claim, err := s.db.GetUser(ctx, req.Login, req.Password)
	if err != nil {
		s.loger.Debug(err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, repository.ErrWrongLoginOrPassword.Error())
		}
		return nil, status.Error(codes.Internal, repository.ErrInternalServer.Error())
	}
	e, err := s.setCookie(ctx, req, claim)
	if err != nil {
		return nil, status.Error(codes.Internal, repository.ErrInternalServer.Error())
	}
	return e, status.Error(codes.OK, repository.ErrSuccessLogined.Error())
}

func (s *publicServer) setCookie(ctx context.Context, req *v1.AuthRequest, claim *repository.Claim) (*v1.Empty, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString([]byte(s.db.SigningKey()))
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
