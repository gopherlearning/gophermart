package rpc

import (
	"context"

	v1 "github.com/gopherlearning/gophermart/proto/v1"
)

type publicServer struct {
	v1.UnimplementedPublicServer
	// db storage.Storage
	// subscriber Subscriber
}

// NewPublicServer возвращает сервер, не требующий авторизации
func NewPublicServer() v1.PublicServer {
	return &publicServer{}
}
func (s *publicServer) UsersRegister(ctx context.Context, req *v1.AuthRequest) (*v1.Empty, error) {
	panic("not implemented") // TODO: Implement
}

func (s *publicServer) UsersLogin(ctx context.Context, req *v1.AuthRequest) (*v1.Empty, error) {
	panic("not implemented") // TODO: Implement
}
