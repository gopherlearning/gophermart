package repository

import (
	"context"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// MockStorage хранилище для тестов
type MockStorage struct {
	Users    map[string]string
	Sessions map[string]*Claim
}

func NewMockStorage() Storage {
	return &MockStorage{
		Users: map[string]string{
			"genry": "12345678",
			"harry": "87654321",
		},
		Sessions: map[string]*Claim{
			"genry": {jwt.RegisteredClaims{Subject: "genry"}},
		},
	}
}
func (m *MockStorage) AccrualMonitor(context.Context, *sync.WaitGroup, string) {
	panic("not implemented") // TODO: Implement
}
func (m *MockStorage) AccrualAdd(string) error {
	panic("not implemented") // TODO: Implement
}
func (m *MockStorage) AccrualGet(string) v1.Order_Status {
	panic("not implemented") // TODO: Implement
}

func (m *MockStorage) GetUserBySession(_ context.Context, _ *Claim) (*Claim, error) {
	panic("not implemented") // TODO: Implement
}
func (m *MockStorage) UnaryCheckToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	panic("not implemented") // TODO: Implement
}
func (m *MockStorage) StreamCheckToken(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	panic("not implemented") // TODO: Implement
}
func (m *MockStorage) SigningKey() string { return "123" }

func (m *MockStorage) CreateOrder(ctx context.Context, id int64) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockStorage) GetBalance(ctx context.Context) (*v1.Balance, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockStorage) GetOrders(ctx context.Context) ([]*v1.Order, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockStorage) CreateWithdraw(ctx context.Context, id string, sum float64) error {
	panic("not implemented") // TODO: Implement
}
func (m *MockStorage) GetWithdrawals(ctx context.Context) ([]*v1.WithdrawRequest, error) {
	panic("not implemented") // TODO: Implement
}
func (m *MockStorage) CreateUser(ctx context.Context, login string, password string) (*Claim, error) {
	if _, ok := m.Users[login]; ok {
		return nil, ErrLoginConflict
	}
	logrus.Info(login, " - ", password)
	if len(password) < 8 {
		return nil, ErrWrongFormat
	}
	s := &Claim{jwt.RegisteredClaims{Subject: "wadwadwadawdawd"}}
	m.Sessions[login] = s
	return s, nil
}

func (m *MockStorage) GetUser(ctx context.Context, login string, password string) (*Claim, error) {
	// if password != m.Users[login] {
	// 	return nil, ErrWrongFormat
	// }
	panic("not implemented") // TODO: Implement
}
