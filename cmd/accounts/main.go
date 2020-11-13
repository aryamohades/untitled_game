package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"untitled_game/accounts/auth"
	"untitled_game/accounts/config"
	"untitled_game/accounts/handler"
	"untitled_game/accounts/register"
	"untitled_game/accounts/session"
	"untitled_game/core/migrate"
	"untitled_game/core/postgres"
)

func main() {
	var flagConfig = flag.String("config", "", "path to config file")
	flag.Parse()

	log := log.New(os.Stdout, "accounts: ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if *flagConfig == "" {
		log.Fatalf("no configuration file was provided. use --config path/to/config.yml to specify a configuration file.")
	}

	cfg, err := config.Load(*flagConfig)
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	if err := migrate.Migrate(cfg.Database.Address); err != nil {
		log.Fatalf("could not perform database migration: %v", err)
	}

	db, err := postgres.Open(postgres.Config{
		Address:         cfg.Database.Address,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetimeSecs * time.Second,
	})
	if err != nil {
		log.Fatalf("could not open database connection: %v", err)
	}
	if err := postgres.Status(db); err != nil {
		log.Fatalf("database status check failed: %v", err)
	}

	sess, err := session.NewStore(session.StoreConfig{
		Redis:      cfg.Sessions.Redis,
		SessionTTL: cfg.Sessions.SessionExpiryMins * time.Minute,
		UserTTL:    cfg.Sessions.UserExpiryMins * time.Minute,
	})

	authService := auth.NewService(sess, auth.NewAccountRepository(db))
	registerService := register.NewService(register.NewAccountRepository(db))

	srv := http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Server.Port),
		Handler:           handler.New(log, sess, authService, registerService),
		ReadTimeout:       cfg.Server.ReadTimeoutSecs * time.Second,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeoutSecs * time.Second,
		WriteTimeout:      cfg.Server.WriteTimeoutSecs * time.Second,
		IdleTimeout:       cfg.Server.IdleTimeoutSecs * time.Second,
	}

	go func() {
		log.Printf("starting server on port: %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve error: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sig := <-shutdown

	log.Printf("shutdown signal received: %v", sig)
	log.Printf("starting graceful server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownGraceSecs*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		srv.Close()
		log.Printf("could not shutdown server gracefully: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("could not close database connection: %v", err)
	}

	if err := sess.Close(); err != nil {
		log.Printf("could not close session store: %v", err)
	}

	log.Println("server shutdown complete, exiting")
}
