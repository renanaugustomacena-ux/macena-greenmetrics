// Command migrate runs goose migrations against TimescaleDB.
//
// Doctrine: Rule 21, Rule 33. ADR: docs/adr/0005-migration-tool-pressly-goose.md.
//
// Usage:
//
//	migrate up
//	migrate up-by-one
//	migrate down
//	migrate status
//	migrate version
//	migrate redo
//	migrate fix
//	migrate create <name> sql
//
// Environment:
//
//	DATABASE_URL — required. Postgres DSN (sslmode=require in production).
//	GOOSE_MIGRATION_DIR — optional override (default: migrations/).
//
// Cross-portfolio invariant: forward-only in production. CI fails any PR that
// edits an applied migration. See docs/SCHEMA-EVOLUTION.md.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	// Blank-import the migrations package so that Go-coded migration init()
	// functions register themselves with goose's global registry. Without
	// this import, only .sql migrations would be applied; the RLS migration
	// 00006_rls_enable.go (Go-coded for per-table policy templating) would
	// silently be skipped.
	_ "github.com/greenmetrics/backend/migrations"
)

const defaultDir = "migrations"

func main() {
	flags := flag.NewFlagSet("migrate", flag.ExitOnError)
	dir := flags.String("dir", envOr("GOOSE_MIGRATION_DIR", defaultDir), "directory containing migration files")
	timeout := flags.Duration("timeout", 10*time.Minute, "max time for a single migration command")
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("flag parse: %v", err)
	}

	args := flags.Args()
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	cmd := args[0]

	dsn := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(dsn) == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, cancel2 := context.WithTimeout(ctx, *timeout)
	defer cancel2()

	if err := run(ctx, db, *dir, cmd, args[1:]); err != nil {
		log.Fatalf("migrate %s: %v", cmd, err)
	}
}

func run(ctx context.Context, db *sql.DB, dir, cmd string, rest []string) error {
	switch cmd {
	case "up":
		return goose.UpContext(ctx, db, dir)
	case "up-by-one":
		return goose.UpByOneContext(ctx, db, dir)
	case "up-to":
		if len(rest) < 1 {
			return errors.New("up-to requires <version>")
		}
		var v int64
		if _, err := fmt.Sscan(rest[0], &v); err != nil {
			return fmt.Errorf("up-to: parse version %q: %w", rest[0], err)
		}
		return goose.UpToContext(ctx, db, dir, v)
	case "down":
		return goose.DownContext(ctx, db, dir)
	case "down-to":
		if len(rest) < 1 {
			return errors.New("down-to requires <version>")
		}
		var v int64
		if _, err := fmt.Sscan(rest[0], &v); err != nil {
			return fmt.Errorf("down-to: parse version %q: %w", rest[0], err)
		}
		return goose.DownToContext(ctx, db, dir, v)
	case "status":
		return goose.StatusContext(ctx, db, dir)
	case "version":
		return goose.VersionContext(ctx, db, dir)
	case "redo":
		return goose.RedoContext(ctx, db, dir)
	case "fix":
		return goose.Fix(dir)
	case "create":
		if len(rest) < 1 {
			return errors.New("create requires <name> [sql|go] (default sql)")
		}
		name := rest[0]
		typ := "sql"
		if len(rest) >= 2 {
			typ = rest[1]
		}
		return goose.Create(db, dir, name, typ)
	default:
		usage()
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: migrate [-dir DIR] [-timeout DURATION] <command> [args]

Commands:
  up                Migrate up to most recent version.
  up-by-one         Migrate up exactly one migration.
  up-to VERSION     Migrate up to (and including) the specified version.
  down              Roll back one migration.
  down-to VERSION   Roll back to (and including) the specified version.
  status            Show migration status.
  version           Print current migration version.
  redo              Roll back and re-apply the most recent migration.
  fix               Renumber out-of-order migration files.
  create NAME [sql|go]  Create a new migration file pair.

Environment:
  DATABASE_URL          Postgres DSN. Required.
  GOOSE_MIGRATION_DIR   Override migration directory (default "migrations").`)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
