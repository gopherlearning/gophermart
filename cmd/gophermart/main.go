package main

import (
	"sync"

	"github.com/gopherlearning/gophermart/cmd/gophermart/config"
	"github.com/gopherlearning/gophermart/cmd/gophermart/server/rpc"
	"github.com/gopherlearning/gophermart/cmd/gophermart/server/web"
	"github.com/gopherlearning/gophermart/cmd/gophermart/storage/postgres"
	"github.com/gopherlearning/gophermart/internal/args"
	"github.com/gopherlearning/gophermart/internal/repository"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	// command-line options:
	// gRPC server endpoint
	wgMain                    = &sync.WaitGroup{}
	wg, loger, addJob, _, err = args.NewApp("gophermarket", "Накопительная система лояльности «Гофермарт»", nil, &config.CLICtl)
)

func main() {
	wgMain.Add(1)
	mainJob := args.StartStopFunc(addJob("Накопительная система лояльности «Гофермарт»"), wgMain)

	var db repository.Storage
	if config.CLICtl.MockStorage {
		db = repository.NewMockStorage()
	} else {
		db, err = postgres.NewStorage(config.CLICtl.DatabaseDSN, loger)
		if err != nil {
			loger.Error(err)
			return
		}
	}
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_auth.StreamServerInterceptor(db.CheckToken),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_auth.UnaryServerInterceptor(db.CheckToken),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	// runtime.WithIncomingHeaderMatcher(),
	// errHandler := localruntime.DefaultHTTPErrorHandler
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(web.DefaultHTTPErrorHandler),
		runtime.WithOutgoingHeaderMatcher(web.HeaderMatcher),
		runtime.WithIncomingHeaderMatcher(web.HeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	wg.Add(2)
	go web.Run(addJob("server web"), wg, config.CLICtl.WebServerAddr, grpcServer, mux, db, nil, loger)
	go rpc.Run(addJob("server grpc"), wg, config.CLICtl.GRPCServerAddr, grpcServer, mux, db, nil, loger)

	wg.Wait()
	mainJob()
}
