package repository

import "errors"

var (
	ErrWrongFormat    = errors.New("неверный формат запроса")
	ErrLoginConflict  = errors.New("логин уже занят")
	ErrInternalServer = errors.New("внутренняя ошибка сервера")
	ErrSessionExpired = errors.New("время жизни сессии истекло")
)
