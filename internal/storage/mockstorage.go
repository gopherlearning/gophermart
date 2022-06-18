package storage

import "github.com/sirupsen/logrus"

// MockStorage хранилище для тестов
type MockStorage struct {
	Users    map[string]string
	Sessions map[string]string
}

func NewMockStorage() Storage {
	return &MockStorage{
		Users: map[string]string{
			"genry": "12345678",
			"harry": "87654321",
		},
		Sessions: map[string]string{
			"genry": "gguygiuohihih",
		},
	}
}

func (m *MockStorage) CreateUser(login string, password string) (*Session, error) {
	if _, ok := m.Users[login]; ok {
		return nil, ErrLoginConflict
	}
	logrus.Info(login, " - ", password)
	if len(password) < 8 {
		return nil, ErrWrongFormat
	}
	s := &Session{AccessToken: "wadwadwadawdawd"}
	m.Sessions[login] = s.AccessToken
	return s, nil
}

func (m *MockStorage) GetUser(login string, password string) (*Session, error) {
	// if password != m.Users[login] {
	// 	return nil, ErrWrongFormat
	// }
	panic("not implemented") // TODO: Implement
}
