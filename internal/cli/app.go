// Package cli реализует интерфейс командной строки GophKeeper.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/al-tokarev/gophkeeper/internal/client"
)

// Client — поверхность API сервера, используемая командами.
type Client interface {
	Login(ctx context.Context, email, password string) error
	Register(ctx context.Context, email, password string) error
	CreateRecord(ctx context.Context, rec client.Record) (*client.Record, error)
	ListRecords(ctx context.Context) ([]client.Record, error)
	GetRecord(ctx context.Context, id string) (*client.Record, error)
	UpdateRecord(ctx context.Context, id string, rec client.Record) (*client.Record, error)
	DeleteRecord(ctx context.Context, id string) (*client.Record, error)
	SetToken(t string)
	Token() string
}

// App содержит зависимости CLI.
type App struct {
	Out    io.Writer
	Err    io.Writer
	Client Client
}

// New создаёт App.
func New(out, errOut io.Writer, c Client) *App {
	return &App{Out: out, Err: errOut, Client: c}
}

// Run разбирает аргументы, выполняет подкоманду и возвращает ошибку.
func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		a.usage()
		return nil
	}

	switch args[0] {
	case "register":
		return a.runRegister(ctx, args[1:])
	case "login":
		return a.runLogin(ctx, args[1:])
	case "list":
		return a.runList(ctx)
	case "get":
		return a.runGet(ctx, args[1:])
	case "add":
		return a.runAdd(ctx, args[1:])
	case "update":
		return a.runUpdate(ctx, args[1:])
	case "delete":
		return a.runDelete(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

// usage печатает справку по командам.
func (a *App) usage() {
	fmt.Fprintln(a.Out, "usage: gophkeeper <command> [flags]")
	fmt.Fprintln(a.Out, "")
	fmt.Fprintln(a.Out, "commands:")
	fmt.Fprintln(a.Out, "  register  register a new user")
	fmt.Fprintln(a.Out, "  login     authenticate and store token")
	fmt.Fprintln(a.Out, "  list      list records")
	fmt.Fprintln(a.Out, "  get       get a record by id")
	fmt.Fprintln(a.Out, "  add       add a record")
	fmt.Fprintln(a.Out, "  update    update a record")
	fmt.Fprintln(a.Out, "  delete    delete a record")
}
