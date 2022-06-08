package config

import (
	"github.com/alecthomas/kong"
	"github.com/posener/complete/cmd/install"
)

type args struct {
	Config               string `help:"Config" default:"config.yaml"`
	Verbose              bool   `short:"v" help:"режим разработчика"`
	WebServerAddr        string `short:"a" help:"адрес WEB сервера" env:"RUN_ADDRESS" default:"127.0.0.1:8090"` // ТЗ
	GRPCServerAddr       string `short:"r" help:"адрес GRPC сервера" env:"RPC_ADDRESS" default:"127.0.0.1:8091"`
	DatabaseDSN          string `short:"d" help:"строка подключения к БД" env:"DATABASE_URI"` // ТЗ
	AccuralSystemAddress string `short:"k" help:"Ключ подписи" env:"ACCRUAL_SYSTEM_ADDRESS"`  // ТЗ
}

func (cli *args) GetConfig() string {
	return cli.Config
}
func (cli *args) InstallCompletionsBash(k *kong.Context) error {
	return install.Install(k.Model.Name)
}
func (cli *args) GetVerbose() bool {
	return cli.Verbose
}

var CLICtl args

type BackendConfig struct {
	Listen struct {
		GRPC string `mapstructure:"grpc" yaml:"grpc"`
		Web  string `mapstructure:"web" yaml:"web"`
	} `mapstructure:"listen" yaml:"listen"`
}
