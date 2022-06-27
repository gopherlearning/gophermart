package repository

import "errors"

var (
	ErrWrongFormat               = errors.New("неверный формат запроса")
	ErrLoginConflict             = errors.New("логин уже занят")
	ErrInternalServer            = errors.New("внутренняя ошибка сервера")
	ErrWrongLoginOrPassword      = errors.New("неверная пара логин/пароль")
	ErrSessionExpired            = errors.New("время жизни сессии истекло")
	ErrNotAuthorized             = errors.New("вы не авторизированы")
	ErrOrderAlreadyUploaded      = errors.New("номер заказа уже был загружен этим пользователем")
	ErrOrderAlreadyUploadedOther = errors.New("номер заказа уже был загружен другим пользователем")
	ErrNoContent                 = errors.New("нет данных для ответа")
	ErrNoWithdrawals             = errors.New("нет ни одного списания")
	ErrLowBalance                = errors.New("на счету недостаточно средств")
	ErrWrongOrderNumber          = errors.New("неверный номер заказа")
)

// not errors
var (
	ErrSuccessRegistered    = errors.New("пользователь успешно зарегистрирован и аутентифицирован")
	ErrSuccessLogined       = errors.New("пользователь успешно зарегистрирован и аутентифицирован")
	ErrSuccessOrderUploaded = errors.New("новый номер заказа принят в обработку")
)
