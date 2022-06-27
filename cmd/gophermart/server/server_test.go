package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"testing"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/publicsuffix"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gopherlearning/gophermart/cmd/gophermart/server/rpc"
	"github.com/gopherlearning/gophermart/cmd/gophermart/server/web"
	"github.com/gopherlearning/gophermart/internal/args"
	"github.com/gopherlearning/gophermart/internal/repository"
)

// func init() {
// 	SetDefaultFailureMode(FailureContinues)

// }
func TestSpec(t *testing.T) {
	SetDefaultFailureMode(FailureContinues)
	defer SetDefaultFailureMode(FailureHalts)
	t.Parallel()

	// Convey("Тестируем сервер", t, func() {

	// Convey("Тестируем публичную часть", func() {
	// 	// jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	// 	// if err != nil {
	// 	// 	log.Fatal(err)
	// 	// }
	// 	// httpClient := &http.Client{
	// 	// 	Jar: jar,
	// 	// }

	// 	public := publicServer{db: storage.NewMockStorage(), loger: logrus.StandardLogger()}
	// 	ses, err := public.UsersRegister(context.Background(), &v1.AuthRequest{Login: "bla", Password: "bla"})
	// 	So(ses, ShouldBeNil)
	// 	So(err, ShouldBeError, status.Error(http.StatusConflict, storage.ErrWrongFormat.Error()))
	// 	ses, err = public.UsersRegister(context.Background(), &v1.AuthRequest{Login: "genry", Password: "bla"})
	// 	So(ses, ShouldBeNil)
	// 	So(err, ShouldBeError, status.Error(http.StatusConflict, storage.ErrLoginConflict.Error()))
	// 	ses, err = public.UsersRegister(context.Background(), &v1.AuthRequest{Login: "puppy", Password: "bla"})
	// 	So(ses, ShouldBeNil)
	// 	So(err, ShouldBeError, status.Error(http.StatusConflict, storage.ErrWrongFormat.Error()))
	// 	ses, err = public.UsersRegister(context.Background(), &v1.AuthRequest{Login: "puppy", Password: "blablabla"})
	// 	So(ses, ShouldNotBeNil)
	// 	So(err, ShouldBeNil)

	// 	// ses, err = public.UsersLogin(context.Background(), &v1.AuthRequest{Login: "puppy", Password: "blablabla"})
	// 	// So(ses, ShouldNotBeNil)
	// 	// So(err, ShouldBeNil)

	// })

	// Convey("Тестируем через http", t, func() {
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
	db := repository.NewMockStorage()
	wg := &sync.WaitGroup{}
	loger := logrus.New()
	loger.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
	loger.SetReportCaller(true)
	wg.Add(2)
	go rpc.Run(context.WithValue(ctx, args.ContextKeyJobName, "server rpc"), wg, "0.0.0.0:7627", grpcServer, mux, db, nil, loger)
	go web.Run(context.WithValue(ctx, args.ContextKeyJobName, "server web"), wg, "127.0.0.1:7628", grpcServer, mux, db, nil, loger)
	time.Sleep(time.Second)
	wgLocal := &sync.WaitGroup{}
	wgLocal.Add(4)
	Convey("Собственные тесты", t, func() {
		defer func() {
			loger.Info("Собственные тесты done")
			wgLocal.Done()
		}()
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		So(err, ShouldBeNil)
		httpClient := &http.Client{
			Jar: jar,
		}
		webURL := "http://127.0.0.1:7628"
		requests := map[string]func() (req *http.Request, status int, response *string, mustErr error){
			"Тест поднятого WEB сервера": func() (req *http.Request, status int, response *string, mustErr error) {
				req, err := http.NewRequest(http.MethodGet, webURL, nil)
				So(err, ShouldBeNil)
				return req, http.StatusNotFound, nil, nil
			},
		}
		for _, f := range requests {
			req, status, response, mustErr := f()
			resp, err := httpClient.Do(req)
			if err != nil {
				So(err, ShouldBeError, mustErr)
			} else {
				So(err, ShouldBeNil)
			}
			if err == nil {
				defer resp.Body.Close()
				So(status, ShouldAlmostEqual, resp.StatusCode)
				body, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				if response != nil {
					So(string(body), ShouldEqual, *response)
				}
			}
		}
	})

	Convey("TestGophermart/TestEndToEnd", t, func() {
		defer func() {
			loger.Error("TestGophermart/TestEndToEnd")
			wgLocal.Done()
		}()
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		So(err, ShouldBeNil)
		httpClient := &http.Client{
			Jar: jar,
		}

		webURL := "http://127.0.0.1:7628"
		requestsEndToEnd := map[string]func() (req *http.Request, status int, response *string, mustErr error){
			"TestGophermart/TestEndToEnd/register_user": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`{"login": "N1WI1XkRhnM","password": "xey57lx6JJn0j8X3saY"}`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/register", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Content-Type", "application/json")
				loger.Info(123)
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestEndToEnd/order_upload": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`205056641066`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/orders", buf)
				So(err, ShouldBeNil)
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestEndToEnd/await_order_processed": func() (req *http.Request, status int, response *string, mustErr error) {
				time.Sleep(time.Second * 5)
				req, err := http.NewRequest(http.MethodGet, webURL+"/api/user/orders", nil)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestEndToEnd/check_balance": func() (req *http.Request, status int, response *string, mustErr error) {
				balance := `{"Current":"729.98","Withdrawn":0}`
				req, err := http.NewRequest(http.MethodGet, webURL+"/api/user/balance", nil)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				return req, http.StatusOK, &balance, nil
			},
			"TestGophermart/TestEndToEnd/withdraw_balance": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`{
								"order": "65627777212855",
									"sum": 700.98
							}`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/balance/withdraw", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestEndToEnd/recheck_balance": func() (req *http.Request, status int, response *string, mustErr error) {
				balance := `{"Current":29,"Withdrawn":"700.98"}`
				req, err := http.NewRequest(http.MethodGet, webURL+"/api/user/balance", nil)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				return req, http.StatusOK, &balance, nil
			},
		}
		for _, f := range requestsEndToEnd {
			req, status, response, mustErr := f()
			resp, err := httpClient.Do(req)
			if err != nil {
				So(err, ShouldBeError, mustErr)
			} else {
				So(err, ShouldBeNil)
			}
			if err == nil {
				defer resp.Body.Close()
				So(status, ShouldAlmostEqual, resp.StatusCode)
				body, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				if response != nil {
					So(string(body), ShouldEqual, *response)
				}
			}
		}
	})
	Convey("TestGophermart/TestUserAuth", t, func() {
		defer func() {
			loger.Error("TestGophermart/TestUserAuth")
			wgLocal.Done()
		}()
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		So(err, ShouldBeNil)
		httpClient := &http.Client{
			Jar: jar,
		}

		webURL := "http://127.0.0.1:7628"
		requestsTestUserAuth := map[string]func() (req *http.Request, status int, response *string, mustErr error){
			"TestGophermart/TestUserAuth/register_user": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`{
								"login": "jMl9g75",
								"password": "iLctBER8ug8ERNkEqaMTbUCBmg"
							}`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/register", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestUserAuth/login_user": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`{
								"login": "jMl9g75",
								"password": "iLctBER8ug8ERNkEqaMTbUCBmg"
							}`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/login", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")
				return req, http.StatusOK, nil, nil
			},
		}
		for name, f := range requestsTestUserAuth {
			loger.Info(name, " started")
			ready := make(chan struct{}, 1)
			loger.Info(name, " started 2")
			Convey(name, func() {
				loger.Info(name, " started 1")
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Recovered. Error:\n", r)
					}
					ready <- struct{}{}
				}()
				req, status, response, mustErr := f()
				resp, err := httpClient.Do(req)
				if err != nil {
					So(err, ShouldBeError, mustErr)
				} else {
					So(err, ShouldBeNil)
				}
				if err == nil {
					defer resp.Body.Close()
					So(status, ShouldAlmostEqual, resp.StatusCode)
					body, err := ioutil.ReadAll(resp.Body)
					So(err, ShouldBeNil)
					if response != nil {
						So(string(body), ShouldEqual, *response)
					}
				}
			})
			loger.Info(name, " started 3")
			<-ready
			loger.Info(name, " tested")
		}
	})
	Convey("TestGophermart/TestUserOrders", t, func() {
		defer func() {
			loger.Error("TestGophermart/TestUserOrders")
			wgLocal.Done()
		}()
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		So(err, ShouldBeNil)
		httpClient := &http.Client{
			Jar: jar,
		}

		webURL := "http://127.0.0.1:7628"
		requestsTestUserOrders := map[string]func() (req *http.Request, status int, response *string, mustErr error){
			"TestGophermart/TestUserOrders/unauthorized_order_upload": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`67728758`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/orders", buf)
				So(err, ShouldBeNil)
				return req, http.StatusUnauthorized, nil, nil
			},
			"TestGophermart/TestUserOrders/unauthorized_orders_list": func() (req *http.Request, status int, response *string, mustErr error) {
				req, err := http.NewRequest(http.MethodGet, webURL+"/api/user/orders", nil)
				So(err, ShouldBeNil)
				return req, http.StatusUnauthorized, nil, nil
			},
			"TestGophermart/TestUserOrders/register_user": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`{"login": "yCc9LI2YE","password": "kdFJcoBJ5gRyMXexJ6SK9h"}`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/register", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestUserOrders/bad_order_upload": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`12345678902`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/orders", buf)
				So(err, ShouldBeNil)
				return req, http.StatusUnprocessableEntity, nil, nil
			},
			"TestGophermart/TestUserOrders/order_upload": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`2808183335`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/orders", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Content-Type", "text/plain")
				return req, http.StatusAccepted, nil, nil
			},
			"TestGophermart/TestUserOrders/duplicate_order_upload_same_user": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`2808183335`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/orders", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Content-Type", "text/plain")
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestUserOrders/orders_list": func() (req *http.Request, status int, response *string, mustErr error) {
				req, err := http.NewRequest(http.MethodGet, webURL+"/api/user/orders", nil)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				// TODO
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestUserOrders/duplicate_order_upload_other_user_register": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`{"login": "MjUYCrwY","password": "ymvUPJiBgH9f1J7zqvflFiskSD"}`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/register", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")
				return req, http.StatusOK, nil, nil
			},
			"TestGophermart/TestUserOrders/duplicate_order_upload_other_user_order": func() (req *http.Request, status int, response *string, mustErr error) {
				buf := bytes.NewBufferString(`2808183335`)
				req, err := http.NewRequest(http.MethodPost, webURL+"/api/user/orders", buf)
				So(err, ShouldBeNil)
				req.Header.Add("Content-Type", "text/plain")
				return req, http.StatusConflict, nil, nil
			},
		}
		for _, f := range requestsTestUserOrders {
			req, status, response, mustErr := f()
			resp, err := httpClient.Do(req)
			if err != nil {
				So(err, ShouldBeError, mustErr)
			} else {
				So(err, ShouldBeNil)
			}
			if err == nil {
				defer resp.Body.Close()
				So(status, ShouldAlmostEqual, resp.StatusCode)
				body, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				if response != nil {
					So(string(body), ShouldEqual, *response)
				}
			}
		}
	})
	wgLocal.Wait()
	cancel()
	wg.Wait()
	// })
	// })
}
