package rpc

import (
	"context"
	"net/http"

	"github.com/gopherlearning/gophermart/internal/storage"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

type publicServer struct {
	v1.UnimplementedPublicServer
	db    storage.Storage
	loger logrus.FieldLogger
	// subscriber Subscriber
}

// NewPublicServer возвращает сервер, не требующий авторизации
func NewPublicServer(db storage.Storage, loger logrus.FieldLogger) v1.PublicServer {
	return &publicServer{db: db, loger: loger}
}
func (s *publicServer) UsersRegister(ctx context.Context, req *v1.AuthRequest) (*v1.Empty, error) {
	session, err := s.db.CreateUser(req.Login, req.Password)
	if err != nil {
		return nil, status.Error(http.StatusConflict, err.Error())
	}
	s.loger.Info(session)
	ctx = context.WithValue(ctx, "grpcgateway-cookie", "SESSIONID=blabla")
	return &v1.Empty{}, nil
}

func (s *publicServer) UsersLogin(ctx context.Context, req *v1.AuthRequest) (*v1.Empty, error) {
	panic("not implemented") // TODO: Implement
}
