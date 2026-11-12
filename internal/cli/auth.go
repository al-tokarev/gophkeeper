package cli

import (
	"context"
	"flag"
	"fmt"
)

// runRegister выполняет команду register.
func (a *App) runRegister(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("register", flag.ContinueOnError)
	fs.SetOutput(a.Err)
	email := fs.String("email", "", "user email")
	password := fs.String("password", "", "user password")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *email == "" || *password == "" {
		return fmt.Errorf("register: --email and --password are required")
	}

	if err := a.Client.Register(ctx, *email, *password); err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "registered: %s\n", *email)
	return nil
}

// runLogin выполняет команду login и сохраняет токен.
func (a *App) runLogin(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(a.Err)
	email := fs.String("email", "", "user email")
	password := fs.String("password", "", "user password")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *email == "" || *password == "" {
		return fmt.Errorf("login: --email and --password are required")
	}

	if err := a.Client.Login(ctx, *email, *password); err != nil {
		return err
	}
	if err := SaveToken(a.Client.Token()); err != nil {
		return fmt.Errorf("login: save token: %w", err)
	}
	fmt.Fprintln(a.Out, "token saved")
	return nil
}
