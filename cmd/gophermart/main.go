package main

import (
	"sync"

	"github.com/gopherlearning/gophermart/cmd/gophermart/config"
	"github.com/gopherlearning/gophermart/cmd/gophermart/server/rpc"
	"github.com/gopherlearning/gophermart/cmd/gophermart/server/web"
	"github.com/gopherlearning/gophermart/internal/args"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	// command-line options:
	// gRPC server endpoint
	wgMain                               = &sync.WaitGroup{}
	wg, loger, addJob, globalCancel, err = args.NewApp("gophermarket", "Накопительная система лояльности «Гофермарт»", nil, &config.CLICtl)
	grpcServer                           = grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
)

func main() {
	wgMain.Add(1)
	mainJob := args.StartStopFunc(addJob("Накопительная система лояльности «Гофермарт»"), wgMain)

	// var db storage.Storage
	// if args.CLICtl.MockStorage {
	// 	db = storage.NewMockStorage()
	// }
	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}))
	// 	MarshalOptions: protojson.MarshalOptions{
	// 		EmitUnpopulated: true,
	// 		UseProtoNames:   false,
	// 	},
	// 	UnmarshalOptions: protojson.UnmarshalOptions{
	// 		DiscardUnknown: false,
	// 	},
	// }))
	wg.Add(2)
	go web.Run(addJob("server web"), wg, config.CLICtl.WebServerAddr, grpcServer, mux, nil, loger)
	go rpc.Run(addJob("server grpc"), wg, config.CLICtl.GRPCServerAddr, grpcServer, mux, nil, loger)

	wg.Wait()
	mainJob()
}
