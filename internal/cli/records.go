package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/al-tokarev/gophkeeper/internal/client"
)

// runList выполняет команду list.
func (a *App) runList(ctx context.Context, args []string) error {
	if a.Client.Token() == "" {
		return fmt.Errorf("not logged in")
	}

	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	search := fs.String("search", "", "metadata search")
	if err := fs.Parse(args); err != nil {
		return err
	}

	recs, err := a.Client.ListRecords(ctx, *search)
	if err != nil {
		if errors.Is(err, client.ErrNetwork) {
			recs = a.Cache.List()
			if len(recs) == 0 {
				return fmt.Errorf("server unavailable and no cached data")
			}
			fmt.Fprintln(a.Err, "warning: server unavailable, showing cached data")
		} else {
			return err
		}
	} else {
		a.Cache.PutAll(recs)
	}

	if len(recs) == 0 {
		fmt.Fprintln(a.Out, "no records")
		return nil
	}
	for _, r := range recs {
		fmt.Fprintf(a.Out, "%s\t%s\t%s\t%s\t%s\n",
			r.ID, r.Type, r.Name, r.MetaData,
			r.UpdatedAt.Format(time.RFC3339))
	}
	return nil
}

// runGet выполняет команду get.
func (a *App) runGet(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.SetOutput(a.Err)
	id := fs.String("id", "", "record id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("get: --id is required")
	}
	if a.Client.Token() == "" {
		return fmt.Errorf("not logged in")
	}

	rec, err := a.Client.GetRecord(ctx, *id)
	if err != nil {
		if errors.Is(err, client.ErrNetwork) {
			cached, ok := a.Cache.Get(*id)
			if !ok {
				return fmt.Errorf("server unavailable and record not in cache")
			}
			fmt.Fprintln(a.Err, "warning: server unavailable, showing cached data")
			rec = &cached
		} else {
			return err
		}
	} else {
		a.Cache.Put(*rec)
	}

	fmt.Fprintf(a.Out, "id:         %s\n", rec.ID)
	fmt.Fprintf(a.Out, "type:       %s\n", rec.Type)
	fmt.Fprintf(a.Out, "name:       %s\n", rec.Name)
	fmt.Fprintf(a.Out, "metadata:   %s\n", rec.MetaData)
	fmt.Fprintf(a.Out, "created_at: %s\n", rec.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(a.Out, "updated_at: %s\n", rec.UpdatedAt.Format(time.RFC3339))
	fmt.Fprintf(a.Out, "data:       %s\n", rec.Data)
	return nil
}

// runAdd выполняет команду add.
func (a *App) runAdd(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(a.Err)
	typ := fs.String("type", "", "record type: login|text|binary|card")
	name := fs.String("name", "", "record name")
	data := fs.String("data", "", "JSON payload")
	file := fs.String("file", "", "file to upload (for --type=binary)")
	meta := fs.String("metadata", "", "metadata")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *typ == "" || *name == "" {
		return fmt.Errorf("add: --type and --name are required")
	}
	if a.Client.Token() == "" {
		return fmt.Errorf("not logged in")
	}

	var payload json.RawMessage

	switch {
	case *file != "":
		if *typ != "binary" {
			return fmt.Errorf("add: --file is only supported for --type=binary")
		}
		if *data != "" {
			return fmt.Errorf("add: --file and --data are mutually exclusive")
		}
		blob, err := readFileAsBase64(*file)
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		payload, err = json.Marshal(map[string]string{"blob": blob})
		if err != nil {
			return fmt.Errorf("add: marshal: %w", err)
		}

	case *data != "":
		if !json.Valid([]byte(*data)) {
			return fmt.Errorf("add: --data must be valid JSON")
		}
		payload = json.RawMessage(*data)

	default:
		return fmt.Errorf("add: --data or --file is required")
	}

	rec, err := a.Client.CreateRecord(ctx, client.Record{
		Type:     *typ,
		Name:     *name,
		Data:     payload,
		MetaData: *meta,
	})
	if err != nil {
		return err
	}
	a.Cache.Put(*rec)
	fmt.Fprintf(a.Out, "created: %s (%s, %s)\n", rec.ID, rec.Type, rec.Name)
	return nil
}

// runUpdate выполняет команду update.
func (a *App) runUpdate(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(a.Err)
	id := fs.String("id", "", "record id")
	name := fs.String("name", "", "record name")
	data := fs.String("data", "", "JSON payload")
	meta := fs.String("metadata", "", "metadata")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" || *name == "" || *data == "" || *meta == "" {
		return fmt.Errorf("update: --id, --name, --data, --metadata are required")
	}
	if !json.Valid([]byte(*data)) {
		return fmt.Errorf("update: --data must be valid JSON")
	}
	if a.Client.Token() == "" {
		return fmt.Errorf("not logged in")
	}

	rec, err := a.Client.UpdateRecord(ctx, *id, client.Record{
		Name:     *name,
		Data:     json.RawMessage(*data),
		MetaData: *meta,
	})
	if err != nil {
		return err
	}
	a.Cache.Put(*rec)
	fmt.Fprintf(a.Out, "updated: %s (%s, %s)\n", rec.ID, rec.Name, rec.UpdatedAt.Format(time.RFC3339))
	return nil
}

// runDelete выполняет команду delete.
func (a *App) runDelete(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(a.Err)
	id := fs.String("id", "", "record id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("delete: --id is required")
	}
	if a.Client.Token() == "" {
		return fmt.Errorf("not logged in")
	}

	rec, err := a.Client.DeleteRecord(ctx, *id)
	if err != nil {
		return err
	}
	a.Cache.Remove(*id)
	if rec.DeletedAt != nil {
		fmt.Fprintf(a.Out, "deleted: %s (%s, %s)\n", rec.ID, rec.Name, rec.DeletedAt.Format(time.RFC3339))
	} else {
		fmt.Fprintf(a.Out, "deleted: %s (%s)\n", rec.ID, rec.Name)
	}
	return nil
}
