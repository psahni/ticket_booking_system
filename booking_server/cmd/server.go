package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"booking_server/db"
	"booking_server/internal/app"
	"booking_server/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var httpServerCommand = &cobra.Command{
	RunE:  runHTTPServer,
	Use:   "http_server",
	Short: "to run http server",
}

func runHTTPServer(_ *cobra.Command, _ []string) error {
	fmt.Println("==runHTTPServer()")
	err := config.Read()
	if err != nil {
		fmt.Println(err)
		panic("Invalid configuration")
	}
	cfg := config.GetConfig()

	gormDB, err := db.Connect()
	if err == nil {
		fmt.Println("Successfully connected to DB", gormDB)
	} else {
		panic("can't connect to DB")
	}

	routes := app.ConfigureRoutes(gormDB)

	ctx := context.Background()

	config.ConnectRedis(cfg.Redis.Host, cfg.Redis.Port)
	client := config.GetRedisClient()

	testRedis(ctx, client)

	fmt.Println("Starting server..")
	ch := make(chan error, 1)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      routes,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Listen for graceful shutdown
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		fmt.Println("Graceful shutdown..")
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}
}

//-----------------------------------------------------------------------------------------

func testRedis(ctx context.Context, client *redis.Client) {
	err := client.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get(ctx, "foo").Result()
	if err != nil {
		panic(err)
	}

	if val == "bar" {
		fmt.Println("redis is working")
	}
}
