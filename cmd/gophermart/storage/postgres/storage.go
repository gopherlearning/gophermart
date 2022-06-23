package postgres

import (
	"context"
	"crypto/rsa"
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
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
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
}

func NewStorage(dsn string, loger logrus.FieldLogger) (repository.Storage, error) {
	connConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	signingKeyRSA, err = jwt.ParseRSAPrivateKeyFromPEM(signingKey)
	if err != nil {
		loger.Error(err)
		return nil, err
	}
	s := &postgresStorage{connConfig: connConfig, loger: loger, maxConnectAttempts: 10}
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

func (s *postgresStorage) CheckToken(ctx context.Context) (context.Context, error) {
	s.loger.Info("авторизация ", ctx)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {

		return nil, status.Error(codes.InvalidArgument, "отсутствуют необходимые заголовки 1")
	}

	if len(md.Get("cookie")) == 0 {
		method, ok := runtime.RPCMethod(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "отсутствуют необходимые заголовки 2")
		}
		switch method {
		case
			"/gopher.market.v1.Public/UsersRegister",
			"/gopher.market.v1.Public/UsersLogin":
			return ctx, nil
		default:
			return nil, status.Error(codes.PermissionDenied, "вы не авторизированы 3")
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
		return nil, status.Error(codes.PermissionDenied, "вы не авторизированы")
	}
	tokenClaim, err := jwt.ParseWithClaims(token, &repository.Claim{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSAPSS); !ok {
			return nil, fmt.Errorf("неверный алгоритм подписи %v", t.Header["alg"])
		}
		return signingKeyRSA, nil
	})
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "вы не авторизированы")
	}
	claim, ok := tokenClaim.Claims.(*repository.Claim)
	if !ok || !tokenClaim.Valid {
		return nil, status.Error(codes.PermissionDenied, "вы не авторизированы")
	}
	c, err := s.GetUserBySession(ctx, claim)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "вы не авторизированы")
	}
	ctx = context.WithValue(ctx, internal.UserID, c.ID)
	return ctx, nil
}

//go:embed signing_key
var signingKey []byte
var signingKeyRSA *rsa.PrivateKey

func (s *postgresStorage) SigningKey() interface{} {
	return signingKeyRSA
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
	// claim. = id
	return claim, nil
}

func (s *postgresStorage) GetUser(ctx context.Context, login string, password string) (*repository.Claim, error) {
	panic("not implemented") // TODO: Implement
}

func (s *postgresStorage) GetUserBySession(ctx context.Context, claim *repository.Claim) (*repository.Claim, error) {
	var userID int64
	err := s.db.QueryRow(ctx, `SELECT user_id FROM sessions WHERE id = $1`, claim.ID).Scan(&userID)
	if err != nil {
		return nil, err
	}
	if claim.ExpiresAt.After(time.Now()) {
		_, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, claim.ID)
		return nil, err
	}
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

// // GetCurrencies ...
// func (s *postgresStorage) GetCurrencies(ctx context.Context) ([]*v1.Currency, error) {
// 	var data []*v1.Currency
// 	// rows, err := s.GetConn(ctx).Query(ctx, `select shortname, name, icon from currencies`)
// 	rows, err := s.GetConn(ctx).Query(ctx, `select shortname, name, icon, type::jsonb from currencies`)
// 	if err != nil {
// 		s.loger.Debug(err)
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var short string
// 		var name string
// 		var icon string
// 		var t v1.Crypto
// 		err = rows.Scan(&short, &name, &icon, &t)
// 		s.loger.Debug(t)
// 		if err != nil {
// 			s.loger.Debug(err)
// 			return nil, err
// 		}
// 		// switch v := t.(type) {
// 		// case *proto.Currency_Card:
// 		// 	s.loger.Info("Card ", name)
// 		// case *proto.Currency_Crypto:
// 		// 	s.loger.Info("Crypto ", name)
// 		// case **proto.Currency_Mobile:
// 		// 	s.loger.Info("Mobile ", name)
// 		// default:

// 		// 	return nil, errors.New(fmt.Sprint("unknown type ", v))
// 		// 	// fmt.Printf("I don't know about type %T!\n", v)
// 		// }
// 		cur := &v1.Currency{
// 			ShortName: short,
// 			Name:      name,
// 			Icon:      icon,
// 			// Type: &t.(proto.),
// 		}
// 		s.loger.Debug(short)
// 		data = append(data, cur)
// 	}
// 	if rows.Err() != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

// // CreateCurrency ...
// func (s *postgresStorage) CreateCurrency(ctx context.Context, c *v1.Currency) error {
// 	_, err := s.GetConn(ctx).Exec(ctx, `INSERT INTO currencies (shortname, name, icon, type) VALUES($1, $2, $3, $4)`, c.ShortName, c.Name, c.Icon, c.GetType())
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // CreateOrder ...
// func (s *postgresStorage) CreateOrder(_ context.Context, _ *v1.Order) (*v1.Order, error) {
// 	panic("not implemented") // TODO: Implement
// }

// // UpdateOrderStatus ...
// func (s *postgresStorage) UpdateOrderStatus(ctx context.Context, orderID string, status *v1.Status) error {
// 	panic("not implemented") // TODO: Implement
// }

// // GetOrder ...
// func (s *postgresStorage) GetOrder(ctx context.Context, orderID string) (*v1.Order, error) {
// 	panic("not implemented") // TODO: Implement
// }

// // GetOrders ...
// func (s *postgresStorage) GetOrders(_ context.Context) ([]*v1.Order, error) {
// 	panic("not implemented") // TODO: Implement
// }

// // GetMessages ...
// func (s *postgresStorage) GetMessages(_ context.Context) ([]*v1.Message, error) {
// 	panic("not implemented") // TODO: Implement
// }

// // CreateMessage ...
// func (s *postgresStorage) CreateMessage(_ context.Context, _ *v1.Message) error {
// 	panic("not implemented") // TODO: Implement
// }
