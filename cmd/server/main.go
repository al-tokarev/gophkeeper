package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/config"
	"github.com/al-tokarev/gophkeeper/internal/cryptoutil"
	"github.com/al-tokarev/gophkeeper/internal/database"
	"github.com/al-tokarev/gophkeeper/internal/database/migrations"
	"github.com/al-tokarev/gophkeeper/internal/handler"
	"github.com/al-tokarev/gophkeeper/internal/logger"
	recordrepository "github.com/al-tokarev/gophkeeper/internal/modules/record/repository"
	recordservice "github.com/al-tokarev/gophkeeper/internal/modules/record/service"
	recordtransport "github.com/al-tokarev/gophkeeper/internal/modules/record/transport"
	userrepository "github.com/al-tokarev/gophkeeper/internal/modules/user/repository"
	userservice "github.com/al-tokarev/gophkeeper/internal/modules/user/service"
	usertransport "github.com/al-tokarev/gophkeeper/internal/modules/user/transport"
	"github.com/al-tokarev/gophkeeper/internal/router"
	"github.com/al-tokarev/gophkeeper/internal/validation"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
)

// main - точка входа в приложение
func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

// run инициализирует и запускает приложение
func run() error {
	err := godotenv.Load()
	if err != nil {
		log.Printf("load .env file: %v", err)
	}

	cfg, err := config.NewConfig(os.Args[1:])
	if err != nil {
		return err
	}

	l, err := logger.NewLogger()
	if err != nil {
		return err
	}

	if err := runMigrations(cfg.DatabaseDSN); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	db, err := database.NewPool(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("error by db connection: %w", err)
	}
	defer db.Close()

	jwtManager, err := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		return fmt.Errorf("error by init jwt: %w", err)
	}
	authManager := auth.NewAuth(jwtManager)

	validate := validation.New()

	cipher, err := cryptoutil.NewCipher(cfg.SecretKey)
	if err != nil {
		return fmt.Errorf("error by cipher create: %w", err)
	}

	userRepository := userrepository.NewRepository(db, l)
	userService := userservice.NewService(userRepository, jwtManager, l)
	userHandler := usertransport.NewHandler(userService, validate, l)

	recRepository := recordrepository.NewRepository(db, l)
	recService := recordservice.NewService(recRepository, cipher, validate, l)
	recHandler := recordtransport.NewHandler(recService, validate, l)

	h := router.NewRouter(handler.NewHandler(userHandler, recHandler), authManager, l)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	l.Infow("Server is starting", "addr", server.Addr)
	return server.ListenAndServe()
}

func runMigrations(dsn string) error {
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
