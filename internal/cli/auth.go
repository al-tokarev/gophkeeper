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
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *email == "" {
		return fmt.Errorf("register: --email is required")
	}

	password, err := readPassword(a.Err, "Password: ")
	if err != nil {
		return fmt.Errorf("register: read password: %w", err)
	}
	if len(password) == 0 {
		return fmt.Errorf("register: password is empty")
	}

	if err := a.Client.Register(ctx, *email, password); err != nil {
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
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *email == "" {
		return fmt.Errorf("login: --email is required")
	}

	password, err := readPassword(a.Err, "Password: ")
	if err != nil {
		return fmt.Errorf("login: read password: %w", err)
	}
	if len(password) == 0 {
		return fmt.Errorf("login: password is empty")
	}

	if err := a.Client.Login(ctx, *email, password); err != nil {
		return err
	}
	if err := SaveToken(a.Client.Token()); err != nil {
		return fmt.Errorf("login: save token: %w", err)
	}
	fmt.Fprintln(a.Out, "token saved")
	return nil
}
