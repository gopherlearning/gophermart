package repository

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/sirupsen/logrus"
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
func (m *MockStorage) GetUserBySession(_ context.Context, _ *Claim) (*Claim, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockStorage) SigningKey() interface{} { return "123" }

func (m *MockStorage) CheckToken(ctx context.Context) (context.Context, error) {

	return ctx, nil
}

func (m *MockStorage) CreateOrder(ctx context.Context, id string) error {
	panic("not implemented") // TODO: Implement
}

func (m *MockStorage) GetOrder(ctx context.Context, id string) (*v1.Order, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockStorage) GetOrders(ctx context.Context, id string) ([]*v1.Order, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MockStorage) Withdrawn(ctx context.Context, id string, sum float64) {
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
