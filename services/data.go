package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charles-d-burton/tailsys/data/sql/migrations"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"go.uber.org/atomic"
	_ "modernc.org/sqlite"
)

type DataManagement struct {
	DB             *sql.DB
	isLocked       atomic.Bool
	MigrationTable string
	DatabaseName   string
	NoTxWrp        bool
}

func (dm *DataManagement) StartDB(dir string) error {
  dbDir := dir + "/db"
	fmt.Println("creating database at: ", dbDir)
	// if err := os.MkdirAll(dir, os.ModePerm); err != nil {
	//   return err
	// }
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite", dbDir+"/tailsys.db")
	if err != nil {
		return err
	}
	dm.DB = db

	fmt.Println("database is created an online")
	fmt.Println("running migrations")
	ctx := context.Background()
	if err := ensureSchema(ctx, db); err != nil {
		return err
	}
	return nil
}

func ensureSchema(ctx context.Context, db *sql.DB) error {
	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		return err
	}

	// List migration sources the provider is aware of.
	fmt.Println("\n=== migration list ===")
	sources := provider.ListSources()
	for _, s := range sources {
		fmt.Printf("%-3s %-2v %v\n", s.Type, s.Version, filepath.Base(s.Path))
	}

	//List status of migrations before applying
	stats, err := provider.Status(ctx)
	if err != nil {
		return err
	}

	fmt.Println("\n=== migration status ===")
	for _, s := range stats {
		fmt.Printf("%-3s %-2v %v\n", s.Source.Type, s.Source.Version, s.State)
	}

	fmt.Println("\n=== log migration output  ===")
	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}

	fmt.Println("\n=== migration results  ===")
	for _, r := range results {
		fmt.Printf("%-3s %-2v done: %v\n", r.Source.Type, r.Source.Version, r.Duration)
	}
	return nil
}
