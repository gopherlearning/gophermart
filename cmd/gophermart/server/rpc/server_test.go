package rpc

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gopherlearning/gophermart/cmd/gophermart/server/web"
	"github.com/gopherlearning/gophermart/internal/args"
	"github.com/gopherlearning/gophermart/internal/storage"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestSpec(t *testing.T) {
	Convey("Тестируем сервер", t, func() {

		Convey("Тестируем публичную часть", func() {
			// jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// httpClient := &http.Client{
			// 	Jar: jar,
			// }

			public := publicServer{db: storage.NewMockStorage(), loger: logrus.StandardLogger()}
			ses, err := public.UsersRegister(context.Background(), &v1.AuthRequest{Login: "bla", Password: "bla"})
			So(ses, ShouldBeNil)
			So(err, ShouldBeError, status.Error(http.StatusConflict, storage.ErrWrongFormat.Error()))
			ses, err = public.UsersRegister(context.Background(), &v1.AuthRequest{Login: "genry", Password: "bla"})
			So(ses, ShouldBeNil)
			So(err, ShouldBeError, status.Error(http.StatusConflict, storage.ErrLoginConflict.Error()))
			ses, err = public.UsersRegister(context.Background(), &v1.AuthRequest{Login: "puppy", Password: "bla"})
			So(ses, ShouldBeNil)
			So(err, ShouldBeError, status.Error(http.StatusConflict, storage.ErrWrongFormat.Error()))
			ses, err = public.UsersRegister(context.Background(), &v1.AuthRequest{Login: "puppy", Password: "blablabla"})
			So(ses, ShouldNotBeNil)
			So(err, ShouldBeNil)

			// ses, err = public.UsersLogin(context.Background(), &v1.AuthRequest{Login: "puppy", Password: "blablabla"})
			// So(ses, ShouldNotBeNil)
			// So(err, ShouldBeNil)

		})

		Convey("Тестируем приватную часть", func() {
			mux := runtime.NewServeMux(
				runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseProtoNames:   true,
						EmitUnpopulated: true,
					},
					UnmarshalOptions: protojson.UnmarshalOptions{
						DiscardUnknown: true,
					},
				}))

			grpcServer := grpc.NewServer(
				grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
					grpc_recovery.StreamServerInterceptor(),
				)),
				grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
					grpc_recovery.UnaryServerInterceptor(),
				)),
			)
			ctx, cancel := context.WithCancel(context.Background())
			db := storage.NewMockStorage()
			wg := &sync.WaitGroup{}
			wg.Add(2)
			go Run(context.WithValue(ctx, args.ContextKeyJobName, "server web"), wg, "0.0.0.0:7627", grpcServer, mux, db, nil, logrus.StandardLogger())
			go web.Run(context.WithValue(ctx, args.ContextKeyJobName, "se web"), wg, "0.0.0.0:7628", grpcServer, mux, nil, logrus.StandardLogger())
			time.Sleep(5 * time.Second)
			cancel()
			wg.Wait()
		})
	})
}
