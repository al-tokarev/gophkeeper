package config

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// Options содержит параметры конфигурации приложения.
type Options struct {
	HTTPAddr    string
	DatabaseDSN string
	SecretKey   []byte
	JWTSecret   []byte
	JWTExpiry   time.Duration
}

// NewConfig загружает конфигурацию из ENV и флагов.
// Приоритет: флаги > ENV > значения по умолчанию.
func NewConfig(args []string) (*Options, error) {
	opts := &Options{
		HTTPAddr:  "localhost:8080",
		JWTExpiry: time.Hour * 24,
	}

	if err := loadFromEnv(opts); err != nil {
		return nil, fmt.Errorf("error by load from env: %w", err)
	}
	if err := parseFlags(opts, args); err != nil {
		return nil, fmt.Errorf("error by parse flags: %w", err)
	}
	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf("error by validate config options: %w", err)
	}

	return opts, nil
}

func loadFromEnv(opts *Options) error {
	if v, ok := os.LookupEnv("HTTP_ADDR"); ok {
		opts.HTTPAddr = v
	}
	if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
		opts.DatabaseDSN = v
	}
	if v, ok := os.LookupEnv("SECRET_KEY"); ok {
		key, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return fmt.Errorf("decode secret key: %w", err)
		}
		opts.SecretKey = key
	}
	if v, ok := os.LookupEnv("JWT_SECRET"); ok {
		opts.JWTSecret = []byte(v)
	}
	if v, ok := os.LookupEnv("JWT_EXPIRY"); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid jwt expiry: %w", err)
		}
		opts.JWTExpiry = d
	}

	return nil
}

func parseFlags(opts *Options, args []string) error {
	fs := flag.NewFlagSet("gophkeeper", flag.ContinueOnError)
	fs.StringVar(&opts.HTTPAddr, "a", opts.HTTPAddr, "HTTP server address")
	fs.StringVar(&opts.DatabaseDSN, "d", opts.DatabaseDSN, "Postgres DSN")
	return fs.Parse(args)
}

func (opts *Options) validate() error {
	if opts.DatabaseDSN == "" {
		return errors.New("database DSN is required (env DATABASE_DSN or flag -d)")
	}
	if len(opts.SecretKey) != 32 {
		return errors.New("incorrect secret key")
	}
	if len(opts.JWTSecret) < 32 {
		return errors.New("incorrect jwt secret")
	}
	if _, port, err := net.SplitHostPort(opts.HTTPAddr); err != nil {
		return fmt.Errorf("incorrect host: %w", err)
	} else if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("incorrect port: %w", err)
	}

	return nil
}
