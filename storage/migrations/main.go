package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"time"
	"x-qdo/jiraclick/pkg/config"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
)

const usageText = `This program runs command on the db. Supported commands are:
  - init - creates version info table in the database
  - up - runs all available migrations.
  - up [target] - runs available migrations up to the target one.
  - down - reverts last migration.
  - reset - reverts all migrations.
  - version - prints current db version.
  - set_version [version] - sets db version without running migrations.
Usage:
  <command> [args]
`

func main() {
	fmt.Printf("starting migrator app\n")
	flag.Usage = usage
	flag.Parse()

	fmt.Printf("Parsing config...\n")
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Parsing DB URL\n")
	opt, err := pg.ParseURL(cfg.Postgres.URL)
	if err != nil {
		panic(fmt.Errorf("failed to connect to Postgres: %s", err.Error()))
	}

	if cfg.Postgres.Insecure {
		opt.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		opt.TLSConfig = nil
	}

	fmt.Printf("Waiting 10s for the network...\n")
	time.Sleep(1 * time.Second)

	fmt.Printf("Connecting to DB...\n")
	db := pg.Connect(opt)
	defer db.Close()
	fmt.Printf("Connected\n")

	fmt.Printf("Testing connection to DB...\n")
	if err := db.Ping(context.Background()); err != nil {
		exitf("failed to ping Postgres: %s", err.Error())
	}

	fmt.Printf("Migrating DB\n")
	oldVersion, newVersion, err := migrations.Run(db, flag.Args()...)
	if err != nil {
		fmt.Printf("Error in DB migration:\n")
		exitf(err.Error())
	}
	if newVersion != oldVersion {
		fmt.Printf("migrated from version %d to %d\n", oldVersion, newVersion)
	} else {
		fmt.Printf("version is %d\n", oldVersion)
	}
}

func usage() {
	fmt.Print(usageText)
	flag.PrintDefaults()
	os.Exit(2)
}

func errorf(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
}

func exitf(s string, args ...interface{}) {
	errorf(s, args...)
	os.Exit(1)
}
