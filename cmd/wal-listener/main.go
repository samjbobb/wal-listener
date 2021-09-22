package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"

	"github.com/ihippik/wal-listener/config"
	"github.com/ihippik/wal-listener/listener"
)

// go build -ldflags "-X main.version=1.0.1" main.go
var version = "0.1.0"

func main() {
	app := &cli.App{
		Name:    "Wal-Listener",
		Usage:   "listen postgres events",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Value:   "config.yml",
				Aliases: []string{"c"},
				Usage:   "path to config file",
			},
		},
		Action: func(c *cli.Context) error {
			cfg, err := getConf(c.String("config"))
			if err != nil {
				logrus.WithError(err).Fatalln("getConf error")
			}
			if err = cfg.Validate(); err != nil {
				logrus.WithError(err).Fatalln("validate config error")
			}

			initLogger(cfg.Logger)

			//sc, err := stan.Connect(cfg.Nats.ClusterID, cfg.Nats.ClientID, stan.NatsURL(cfg.Nats.Address))
			//if err != nil {
			//	logrus.WithError(err).Fatalln(listener.ErrNatsConnection)
			//}

			conn, rConn, err := initPgxConnections(cfg.Database)
			if err != nil {
				logrus.Fatal(err)
			}
			repo := listener.NewRepository(conn)
			//natsPublisher := listener.NewNatsPublisher(sc)
			parser := listener.NewBinaryParser(binary.BigEndian)
			service := listener.NewWalListener(cfg, repo, rConn, parser)
			return service.Process()
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

// getConf load config from file.
func getConf(path string) (*config.Config, error) {
	var cfg config.Config
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %w", err)
	}

	return &cfg, nil
}
