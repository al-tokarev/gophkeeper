package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/al-tokarev/gophkeeper/internal/client"
)

// runList выполняет команду list.
func (a *App) runList(ctx context.Context) error {
	if a.Client.Token() == "" {
		return fmt.Errorf("not logged in")
	}

	recs, err := a.Client.ListRecords(ctx)
	if err != nil {
		return err
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
		return err
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
	meta := fs.String("metadata", "", "metadata")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *typ == "" || *name == "" || *data == "" {
		return fmt.Errorf("add: --type, --name, --data are required")
	}
	if !json.Valid([]byte(*data)) {
		return fmt.Errorf("add: --data must be valid JSON")
	}
	if a.Client.Token() == "" {
		return fmt.Errorf("not logged in")
	}

	rec, err := a.Client.CreateRecord(ctx, client.Record{
		Type:     *typ,
		Name:     *name,
		Data:     json.RawMessage(*data),
		MetaData: *meta,
	})
	if err != nil {
		return err
	}
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
	if rec.DeletedAt != nil {
		fmt.Fprintf(a.Out, "deleted: %s (%s, %s)\n", rec.ID, rec.Name, rec.DeletedAt.Format(time.RFC3339))
	} else {
		fmt.Fprintf(a.Out, "deleted: %s (%s)\n", rec.ID, rec.Name)
	}
	return nil
}
