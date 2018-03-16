// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var (
	ctx context.Context
	db  *sql.DB
)

func ExampleDB_Query() error {
	age := 27
	rows, err := db.Query("SELECT name FROM users WHERE age=?;", age)
	if err != nil {
		return err
	}

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			break
		}
		fmt.Printf("%s is %d\n", name, age)
	}
	// If the database is being written to ensure to check for Close
	// errors that may be returned from the driver. The query may
	// encounter an auto-commit error and be forced to rollback changes.
	rerr := rows.Close()
	if rerr != nil {
		return err
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func ExampleDB_QueryRow() error {
	id := 123
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id=?;", id).Scan(&username)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("No user with that ID.")
		return nil
	case err != nil:
		fmt.Println("Query error: ", err)
		return err
	default:
		fmt.Printf("Username is %s\n", username)
		return nil
	}
}

func ExampleDB_Query_multipleResultSets() error {
	age := 27
	q := `
create temp table uid (id bigint); -- Create temp table for queries.
insert into uid
select id from users where age < ?; -- Populate temp table.

-- First result set.
select
	users.id, name
from
	users
	join uid on users.id = uid.id
;

-- Second result set.
select 
	ur.user, ur.role
from
	user_roles as ur
	join uid on uid.id = ur.user
;
	`
	rows, err := db.Query(q, age)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id   int64
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		fmt.Printf("id %d name is %s\n", id, name)
	}
	if !rows.NextResultSet() {
		return fmt.Errorf("expected more result sets: %v", rows.Err())
	}
	var roleMap = map[int64]string{
		1: "user",
		2: "admin",
		3: "gopher",
	}
	for rows.Next() {
		var (
			id   int64
			role int64
		)
		if err := rows.Scan(&id, &role); err != nil {
			return err
		}
		fmt.Printf("id %d has role %s\n", id, roleMap[role])
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func ExampleDB_PingContext() {
	// Ping and PingContext may be used to determine if communication with
	// the database server is still possible.
	//
	// When used in a command line application Ping may be used to establish
	// that further queries are possible; that the provided DSN is valid.
	//
	// When used in long running service Ping may be part of the health
	// checking system.

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	status := "up"
	if err := db.PingContext(ctx); err != nil {
		status = "down"
	}
	fmt.Println(status)
}
