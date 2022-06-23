package web

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/gopherlearning/gophermart/internal"
	"github.com/gopherlearning/gophermart/internal/args"
	"github.com/gopherlearning/gophermart/internal/repository"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, wg *sync.WaitGroup, listen string, grpcServer *grpc.Server, mux *runtime.ServeMux, db repository.Storage, tlsConfig *tls.Config, loger logrus.FieldLogger) {

	onStop := args.StartStopFunc(ctx, wg)
	defer onStop()

	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(resp, req.WithContext(context.WithValue(req.Context(), internal.PathKey, req.URL.Path)))
	})
	e := echo.New()
	e.HideBanner = true
	// echologrus.SetLoger(loger)
	// e.Logger = echologrus.GetEchoLogger()
	// e.Use(echologrus.Hook())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, err := db.CheckToken(c.Request().Context())
			if err != nil {
				return c.HTML(http.StatusForbidden, err.Error())
			}
			loger.Info(c.Request().Context().Value(internal.UserID))
			return next(c)
		}
	})
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return echo.WrapHandler(handler)
	})
	// Start server
	go func() {
		if tlsConfig != nil {
			tlsServer := http.Server{
				Addr:      listen,
				TLSConfig: tlsConfig,
			}
			if err := e.StartServer(&tlsServer); err != nil {
				loger.Info("Shutting down the server")
			}
			return
		}
		if err := e.Start(listen); err != nil && err != http.ErrServerClosed {
			loger.Error("web server stoped with error: ", err)
		}
	}()
	<-ctx.Done()
	loger.Info("server/web: stoping")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		loger.Warn(err)
	}
}
