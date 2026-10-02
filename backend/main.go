package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Subcommands for one-off admin tasks; the default is to serve.
	if len(os.Args) > 1 {
		cmds := map[string]func(Config, []string) error{"claim-legacy": claimLegacy, "backup": backup}
		if cmd, ok := cmds[os.Args[1]]; ok {
			if err := cmd(cfg, os.Args[2:]); err != nil {
				log.Fatal(err)
			}
			return
		}
	}

	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address")
	flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "path to the SQLite database")
	flag.StringVar(&cfg.LegacyDB, "legacy-json", cfg.LegacyDB, "old JSON store to import once, if present")
	flag.Parse()

	store, err := OpenStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.Close()
	if n, err := store.ImportLegacyJSON(context.Background(), cfg.LegacyDB); err != nil {
		log.Fatalf("import %s: %v", cfg.LegacyDB, err)
	} else if n > 0 {
		log.Printf("imported %d sessions from %s — run `smistudy-api claim-legacy -email you@example.com` to attach them to your account", n, cfg.LegacyDB)
	}

	go purgeLoop(store)

	app := NewApp(cfg, store, NewMailer(cfg))
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	// Finish in-flight requests on SIGTERM so restarts behind a load balancer drop nothing.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
	}()

	log.Printf("smistudy api listening on %s (db: %s, public url: %s, google sign-in: %v)",
		cfg.Addr, cfg.DBPath, cfg.PublicURL, cfg.GoogleEnabled())
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func purgeLoop(store *Store) {
	for ; ; time.Sleep(time.Hour) {
		if err := store.PurgeExpired(context.Background()); err != nil {
			log.Printf("purge expired sessions: %v", err)
		}
	}
}

// claimLegacy attaches sessions from before accounts existed to one user.
func claimLegacy(cfg Config, args []string) error {
	fs := flag.NewFlagSet("claim-legacy", flag.ExitOnError)
	email := fs.String("email", "", "email of the account that should own the old sessions")
	db := fs.String("db", cfg.DBPath, "path to the SQLite database")
	fs.Parse(args)

	store, err := OpenStore(*db)
	if err != nil {
		return err
	}
	defer store.Close()
	ctx := context.Background()
	addr, err := normalizeEmail(*email)
	if err != nil {
		return err
	}
	u, err := store.UserByEmail(ctx, addr)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("no account with email %s — sign up first", addr)
	}
	n, err := store.ClaimLegacy(ctx, u.ID)
	if err != nil {
		return err
	}
	fmt.Printf("gave %d old sessions to %s\n", n, u.Username)
	return nil
}

// backup writes a consistent copy of the live database (safe while the API is running).
func backup(cfg Config, args []string) error {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	db := fs.String("db", cfg.DBPath, "path to the SQLite database")
	out := fs.String("out", "", "file to write the backup to (must not exist)")
	fs.Parse(args)
	if *out == "" {
		return errors.New("-out is required")
	}
	store, err := OpenStore(*db)
	if err != nil {
		return err
	}
	defer store.Close()
	if err := store.Backup(context.Background(), *out); err != nil {
		return err
	}
	fmt.Printf("backed up %s to %s\n", *db, *out)
	return nil
}
