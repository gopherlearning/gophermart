package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gopherlearning/gophermart/cmd/gophermart/storage/postgres/migrations"
	"github.com/gopherlearning/gophermart/internal"
	"github.com/gopherlearning/gophermart/internal/migrate"
	"github.com/gopherlearning/gophermart/internal/repository"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type postgresStorage struct {
	mu                 sync.Mutex
	db                 *pgxpool.Pool
	connConfig         *pgxpool.Config
	loger              logrus.FieldLogger
	maxConnectAttempts int
	secretKey          string
}

func NewStorage(dsn string, loger logrus.FieldLogger, secretKey string) (repository.Storage, error) {
	connConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	s := &postgresStorage{connConfig: connConfig, loger: loger, maxConnectAttempts: 10, secretKey: secretKey}
	err = migrate.MigrateFromFS(context.Background(), s.GetConn(context.Background()), &migrations.Migrations, loger)
	if err != nil {
		loger.Error(err)
		return nil, err
	}
	return s, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// func unaryInterceptor(ctx context.Context,
// 	req interface{},
// 	info *grpc.UnaryServerInfo,
// 	handler grpc.UnaryHandler,
// ) (interface{}, error) {
// 	return handler(ctx, req)
// }

func (s *postgresStorage) StreamCheckToken(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	s.loger.Info("авторизация ", ss.Context())
	s.loger.Println("--> unary interceptor: ", info.FullMethod)
	s.loger.Println("--> unary interceptor: ", info.FullMethod)
	userID, err := s.checkToken(ss.Context(), info.FullMethod)
	if err != nil {
		return err
	}
	ss.SetHeader(metadata.Pairs(fmt.Sprint(internal.ContextKeyUserID{}), userID))
	return handler(srv, ss)
}

func (s *postgresStorage) UnaryCheckToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	s.loger.Info("авторизация ", ctx)
	s.loger.Println("--> unary interceptor: ", info.FullMethod)
	s.loger.Println("--> unary interceptor: ", info.FullMethod)
	userID, err := s.checkToken(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}
	return handler(context.WithValue(ctx, internal.ContextKeyUserID{}, userID), req)

}

func (s *postgresStorage) checkToken(ctx context.Context, method string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.InvalidArgument, "отсутствуют необходимые заголовки 1.1")
	}
	s.loger.Info(md)
	if len(md.Get("cookie")) == 0 {
		switch method {
		case
			"/gopher.market.v1.Public/UsersRegister",
			"/gopher.market.v1.Public/UsersLogin":
			return "", nil
		default:
			return "", status.Error(codes.PermissionDenied, "вы не авторизированы 3")
		}
	}
	var token string
	for _, v := range md.Get("cookie") {
		if strings.Split(v, "=")[0] != "accesstoken" {
			continue
		}
		token = strings.Split(v, "=")[1]
	}
	if len(token) == 0 {
		return "", status.Error(codes.PermissionDenied, "вы не авторизированы 4")
	}
	tokenClaim, err := jwt.ParseWithClaims(token, &repository.Claim{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный алгоритм подписи %v", t.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return "", status.Error(codes.PermissionDenied, "вы не авторизированы 5 "+err.Error())
	}
	claim, ok := tokenClaim.Claims.(*repository.Claim)
	if !ok || !tokenClaim.Valid {
		return "", status.Error(codes.PermissionDenied, "вы не авторизированы 6")
	}
	c, err := s.GetUserBySession(ctx, claim)
	if err != nil {
		switch err {
		case pgx.ErrNoRows, repository.ErrSessionExpired:
			return "", nil
		default:
			return "", status.Error(codes.PermissionDenied, "вы не авторизированы 7 "+err.Error())
		}
	}
	switch method {
	case "/gopher.market.v1.Public/UsersRegister":
		return "", status.Error(codes.PermissionDenied, "вы уже зарегистрированы")
	case "/gopher.market.v1.Public/UsersLogin":
		return "", status.Error(codes.PermissionDenied, "вы уже авторизированы")
	default:
		return c.ID, nil
	}
}

func (s *postgresStorage) SigningKey() string {
	return s.secretKey
}
func (s *postgresStorage) CreateUser(ctx context.Context, login string, password string) (*repository.Claim, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	var id int64
	tx, err := s.GetConn(ctx).BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(ctx, `INSERT INTO users (login, hashed_password) VALUES($1, $2) RETURNING (id)`, login, hashedPassword).Scan(&id)
	if err != nil {
		return nil, err
	}
	s.loger.Debug(id)
	claim := &repository.Claim{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(id)),
			Audience:  jwt.ClaimStrings{login},
			Issuer:    "CreateUser",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 20)),
		},
	}
	err = tx.QueryRow(ctx, `INSERT INTO sessions (user_id, claim) VALUES($1, $2) RETURNING (id)`, id, claim).Scan(&id)
	if err != nil {
		return nil, err
	}
	s.loger.Debug(id)
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	claim.ID = fmt.Sprint(id)
	return claim, nil
}

func (s *postgresStorage) GetUser(ctx context.Context, login string, password string) (*repository.Claim, error) {
	panic("not implemented") // TODO: Implement
}

func (s *postgresStorage) GetUserBySession(ctx context.Context, claim *repository.Claim) (*repository.Claim, error) {
	var userID int64
	s.loger.Info(claim.ID)
	s.loger.Infof("%+v", claim)
	searchID, err := strconv.Atoi(claim.ID)
	if err != nil {
		return nil, err
	}
	s.loger.Info(searchID)
	err = s.GetConn(ctx).QueryRow(ctx, `SELECT user_id FROM sessions WHERE id = $1`, searchID).Scan(&userID)
	if err != nil {
		return nil, err
	}
	s.loger.Info(searchID)
	s.loger.Info(claim.ExpiresAt)
	if claim.ExpiresAt != nil && claim.ExpiresAt.Before(time.Now()) {
		_, err = s.GetConn(ctx).Exec(ctx, `DELETE FROM sessions WHERE id = $1`, searchID)
		return nil, fmt.Errorf("session expired")
	}
	s.loger.Info(searchID)
	return claim, nil
}

func (s *postgresStorage) CreateOrder(ctx context.Context, id string) error {
	panic("not implemented") // TODO: Implement
}

func (s *postgresStorage) GetOrder(ctx context.Context, id string) (*v1.Order, error) {
	panic("not implemented") // TODO: Implement
}

func (s *postgresStorage) GetOrders(ctx context.Context, id string) ([]*v1.Order, error) {
	panic("not implemented") // TODO: Implement
}

func (s *postgresStorage) Withdrawn(ctx context.Context, id string, sum float64) {
	panic("not implemented") // TODO: Implement
}

// Close ...
func (s *postgresStorage) Close(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	s.db.Close()
	return nil
}

func (s *postgresStorage) reconnect(ctx context.Context) (*pgxpool.Pool, error) {

	pool, err := pgxpool.ConnectConfig(context.Background(), s.connConfig)

	if err != nil {
		return nil, fmt.Errorf("unable to connection to database: %v", err)
	}
	if err = pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("couldn't ping postgre database: %v", err)
	}
	return pool, err
}

func (s *postgresStorage) GetConn(ctx context.Context) *pgxpool.Pool {
	var err error

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil || s.db.Ping(ctx) != nil {
		attempt := 0
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if attempt >= s.maxConnectAttempts {
				s.loger.Errorf("connection failed after %d attempt\n", attempt)
			}
			attempt++

			s.loger.Info("reconnecting...")

			s.db, err = s.reconnect(ctx)
			if err == nil {
				ticker = time.NewTicker(3 * time.Second)
				return s.db
			}

			s.loger.Errorf("connection was lost. Error: %s. Waiting for 5 sec...\n", err)
		}
		return nil
	}
	return s.db
}

// Ping ...
func (s *postgresStorage) Ping(ctx context.Context) error {
	ctx_, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ping := make(chan error)
	go func() {
		ping <- s.GetConn(ctx_).Ping(ctx_)
	}()
	select {
	case err := <-ping:
		return err
	case <-ctx_.Done():
		return fmt.Errorf("context closed")
	}

}
