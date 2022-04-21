package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/SYSTEMTerror/GoWallet/internal/app"
	"github.com/SYSTEMTerror/GoWallet/internal/pkg/wallet"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

func main() {
	viper.AddConfigPath("../config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	host := viper.Get("server.host").(string)
	port := viper.Get("server.port").(string)
	dsn := viper.Get("database.dsn").(string)
	secretKey := viper.Get("security.secret_key").(string)

	if err := execute(host, port, dsn, secretKey); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(host string, port string, dsn string, secretKey string) (err error) {
	deps := []interface{}{
		app.NewServer,
		func() (string) {
			return secretKey
		},
		mux.NewRouter,
		func() (*pgxpool.Pool, error) {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			return pgxpool.Connect(ctx, dsn)
		},
		wallet.NewService,
		func(server *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(host, port),
				Handler: server,
			}
		},
	}

	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			return err
		}
	}

	err = container.Invoke(func(server *app.Server) {
		server.Init()
	})
	if err != nil {
		return err
	}

	return container.Invoke(func(server *http.Server) error {
		return server.ListenAndServe()
	})
}
