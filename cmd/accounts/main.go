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
	"untitled_game/core/envoy"
	"untitled_game/core/migrate"
	"untitled_game/core/postgres"
	"untitled_game/proto"
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

	env, err := envoy.New(envoy.Config{
		Redis:   cfg.Envoy.Redis,
		Service: proto.ServiceAccounts,
	})
	if err != nil {
		log.Fatalf("could not create envoy: %v", err)
	}

	// DEMO of inter-service communication using envoy
	envSvc := exampleService{log, env}
	go func() {
		if err := env.Receive(proto.RouteGetAccount, envSvc.exampleReceiver); err != nil {
			log.Printf("envoy receive error (route=%d): %v", proto.RouteGetAccount, err)
		}
	}()

	// EXAMPLE of using envoy to send request to service and wait for response
	go func() {
		for {
			// Wait 2 seconds between requests
			time.Sleep(2 * time.Second)

			// Send a simple "ping" message
			r := envoy.Request{
				Service: proto.ServiceAccounts,
				Route:   proto.RouteGetAccount,
				Data:    "ping",
			}

			// Send the request and save the returned id
			id, err := env.Send(r)
			if err != nil {
				log.Printf("envoy send error: %v", err)
				continue
			}

			// Use the unique id of the request to wait for a response with a 5 second timeout
			res, err := env.Wait(id, 5)
			if err != nil {
				log.Printf("envoy wait error: %v", err)
				continue
			}

			// Log the response
			log.Printf("envoy response: %v", res)
		}
	}()
	// END DEMO envoy

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

type exampleService struct {
	log *log.Logger
	env *envoy.Envoy
}

func (s *exampleService) exampleReceiver(data interface{}, res envoy.Responder) {
	err := s.env.Receive(proto.RouteGetAccount, func(data interface{}, res envoy.Responder) {
		s.log.Printf("received data: %v", data)

		r := envoy.Response{
			Data:       "pong",
			ExpirySecs: 5,
		}
		if err := res(r); err != nil {
			s.log.Printf("envoy respond error: %v", err)
		}
	})

	if err != nil {
		s.log.Printf("envoy receive error (service route = %d): %v", 1, err)
	}
}
