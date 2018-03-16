// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func Example_openDBCLI() {
	id := flag.Int64("id", 0, "person ID to find")
	dsn := flag.String("dsn", os.Getenv("DSN"), "connection data source name")
	flag.Parse()

	if len(*dsn) == 0 {
		log.Fatal("missing dsn flag")
	}
	if *id == 0 {
		log.Fatal("missing person ID")
	}
	// Opening a driver typically will not attempt to connect to the database.
	db, err := sql.Open("driver-name", *dsn)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal("unable to use data source name", err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(3)

	appBaseCtx, stop := context.WithCancel(context.Background())
	defer stop()

	appSignal := make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)

	go func() {
		select {
		case <-appSignal:
			stop()
		}
	}()

	// Check if the DSN provided by the user is valid and the server accessible.
	ctx, cancel := context.WithTimeout(appBaseCtx, 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("unable to connect to database for DSN %q: %v", *dsn, err)
	}

	ctx, cancel = context.WithTimeout(appBaseCtx, 5*time.Second)
	defer cancel()

	// Attempt to run the query.
	var name string
	err = db.QueryRowContext(ctx, "select p.name from people as p where p.id = :id;", sql.Named("id", *id)).Scan(&name)
	if err != nil {
		log.Fatal("unable to execute search query", err)
	}
	fmt.Println("name=", name)
}
