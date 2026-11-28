// Command gophkeeper — клиент командной строки GophKeeper.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/al-tokarev/gophkeeper/internal/cli"
	"github.com/al-tokarev/gophkeeper/internal/client"
)

var (
	buildVersion = "unknown"
	buildDate    = "unknown"
)

// serverURL возвращает адрес сервера из переменной окружения GOPHKEEPER_SERVER,
// либо https://localhost:8080 по умолчанию.
func serverURL() string {
	if s := os.Getenv("GOPHKEEPER_SERVER"); s != "" {
		return strings.TrimRight(s, "/")
	}
	return "https://localhost:8080"
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Printf("gophkeeper %s (built %s)\n", buildVersion, buildDate)
		return
	}

	c, err := client.New(serverURL())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	c.SetToken(cli.LoadToken())

	app := cli.New(os.Stdout, os.Stderr, c)
	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
