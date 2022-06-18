package storage

type Session struct {
	AccessToken string
}

type Storage interface {
	CreateUser(login, password string) (*Session, error)
	GetUser(login, password string) (*Session, error)
}
