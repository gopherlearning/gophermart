package args

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/caarlos0/env"
	"github.com/creasty/defaults"
	"github.com/posener/complete"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/willabides/kongplete"
	"gopkg.in/yaml.v2"
)

var (
	CTX           *kong.Context
	Loger         logrus.FieldLogger
	appNameMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app",
		Help: "Application name",
	}, []string{"name"})
	metrics   = []prometheus.Collector{}
	isHealthy = &atomic.Value{}
	isReady   = &atomic.Value{}
	AddJob    func(name string) context.Context
)

// RegisterMetric .
func RegisterMetric(v prometheus.Collector) {
	metrics = append(metrics, v)
	prometheus.Register(v)
}

type CLI interface {
	GetVerbose() bool
	InstallCompletionsBash(k *kong.Context) error
	GetConfig() string
}

// NewApp .
func NewApp(name, desc string, cfg interface{}, cli CLI) (
	wg *sync.WaitGroup,
	loger *logrus.Logger,
	addJob func(name string) context.Context,
	globalCancel context.CancelFunc,
	err error,
) {
	var globalCtx context.Context
	globalCtx, globalCancel = context.WithCancel(context.Background())

	addJob = func(n string) context.Context {
		return context.WithValue(globalCtx, ContextKeyJobName, n)
	}
	AddJob = addJob
	// Create a kong parser as usual, but don't run Parse quite yet.
	// только для прохождения теста
	for i := 0; i < len(os.Args); i++ {
		if strings.Contains(os.Args[i], "=") {
			a := strings.Split(os.Args[i], "=")
			if a[0] == "-r" {
				os.Args[i] = fmt.Sprintf("--restore=%s", a[1])
				continue
			}
			if a[0] == "-d" {
				os.Args[i] = fmt.Sprintf("--database-dsn=%s", a[1])
				continue
			}
			os.Args = append(os.Args[:i], append(a, os.Args[i+1:]...)...)
		}
	}
	parser := kong.Must(cli,
		kong.Name(name),
		kong.Description(desc),
		kong.UsageOnError(),
	)
	err = env.Parse(cli)
	if err != nil {
		logrus.Fatal(err)
	}
	// Run kongplete.Complete to handle completion requests
	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)
	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	loger = logrus.StandardLogger()
	switch ctx.Command() {
	default:
		// fmt.Println(cli.Run.Verbose)
	case "completion bash":
		Loger.Info("installing...")
		err := cli.InstallCompletionsBash(ctx)
		if err != nil {
			Loger.Error(err)
		}
		os.Exit(0)
	}
	CTX = ctx
	// kong.Parse(&CLI)
	if cli.GetVerbose() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true)
		logrus.SetFormatter(&logrus.TextFormatter{})
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	Loger = logrus.StandardLogger()

	loger.Debug(ctx.Args)
	appNameMetric.WithLabelValues(os.Getenv("APP")).Set(1)
	terminate := make(chan os.Signal)
	wg = &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-terminate
		globalCancel()
		go func() {
			time.Sleep(15 * time.Second)
			logrus.Error("Program exit by timeout")
			os.Exit(1)
		}()
	}()

	// wg.Add(1)
	// go startMetrics(AddJob("metrics"), wg, metrics)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM)
	if cfg != nil {
		if err := readFromFile(cli.GetConfig(), cfg); err != nil {
			return nil, loger, nil, nil, err
		}
	}
	return
}

// StartStopFunc .
func StartStopFunc(ctx context.Context, wg *sync.WaitGroup) func() {
	job := ctx.Value(ContextKeyJobName).(string)
	if len(job) != 0 {
		logrus.Info(job + " started")
	}
	return func() {
		wg.Done()
		logrus.Info(job + " stoped")
	}
}

func readFromFile(filename string, out interface{}) error {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		yamlData, err := yaml.Marshal(out)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filename, yamlData, 0600)
		if err != nil {
			return err
		}
	}
	if err := defaults.Set(out); err != nil {
		return err
	}

	viper.AutomaticEnv()
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		logrus.Error(err)
		return err
	}

	if err := viper.Unmarshal(out); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func metricPort() string {
	port := os.Getenv("METRICS")
	if len(port) != 0 && strings.ContainsRune(port, ':') {
		return port
	}
	return ":9100"
}

// StartMetrics .
func startMetrics(ctx context.Context, wg *sync.WaitGroup, metrics []prometheus.Collector) error {
	onStop := StartStopFunc(ctx, wg)
	defer onStop()
	isHealthy.Store(false)
	isReady.Store(false)
	for _, v := range metrics {
		prometheus.Register(v)
	}
	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			switch req.URL.Path {
			case "/readyz":
				if isReady == nil || !isReady.Load().(bool) {
					resp.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				resp.WriteHeader(http.StatusOK)
				return
			case "/healthz":
				if isHealthy == nil || !isHealthy.Load().(bool) {
					resp.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				resp.WriteHeader(http.StatusOK)
				return
			}
		}
		promhttp.Handler().ServeHTTP(resp, req)
	})
	h := &http.Server{Addr: metricPort(), Handler: handler}
	go func() {
		err := h.ListenAndServe()
		if err != nil {
			if err.Error() != "http: Server closed" {
				logrus.WithField("error", err).Warn("start metrics failed")
			}
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return h.Shutdown(ctx)
}

func SetReady(r bool) {
	isReady.Store(r)
}

func SetHealthy(h bool) {
	isHealthy.Store(h)
}

type contextKey string

func (c contextKey) String() string {
	return "mypackage context key " + string(c)
}

var (
	// ContextKeyJobName .
	ContextKeyJobName = contextKey("jobName")
)
