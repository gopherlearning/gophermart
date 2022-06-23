package rpc

import (
	"context"
	"crypto/tls"
	_ "embed"
	"net"
	"sync"

	"github.com/gopherlearning/gophermart/cmd/gophermart/config"
	"github.com/gopherlearning/gophermart/internal/args"
	"github.com/gopherlearning/gophermart/internal/repository"
	v1 "github.com/gopherlearning/gophermart/proto/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(ctx context.Context, wg *sync.WaitGroup, listen string, grpcServer *grpc.Server, mux *runtime.ServeMux, db repository.Storage, tlsConfig *tls.Config, loger logrus.FieldLogger) {
	onStop := args.StartStopFunc(ctx, wg)
	defer onStop()
	public := NewPublicServer(db, loger)
	private := NewPrivateServer()
	err := v1.RegisterPrivateHandlerFromEndpoint(ctx, mux, config.CLICtl.GRPCServerAddr, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		loger.Fatal(err)
		return
	}
	// v1.RegisterPublicHandlerClient(ctx, mux)
	err = v1.RegisterPublicHandlerFromEndpoint(ctx, mux, config.CLICtl.GRPCServerAddr, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		loger.Fatal(err)
		return
	}
	v1.RegisterPublicServer(grpcServer, public)
	v1.RegisterPrivateServer(grpcServer, private)
	go func() {
		lis, err := net.Listen("tcp", listen)
		if err != nil {
			loger.Error(err)
		}
		if tlsConfig != nil {
			lis = tls.NewListener(lis, tlsConfig)
		}
		defer lis.Close()
		if err := grpcServer.Serve(lis); err != nil {
			loger.Error(err)
			args.SetHealthy(false)
		}
	}()
	<-ctx.Done()
	// subscriber.UnsubscribeAll()
	defer func() {
		if r := recover(); r != nil {
			loger.Warnf("server/rpc: stoped panic - %v", r)
		}
	}()
	grpcServer.GracefulStop()
}
