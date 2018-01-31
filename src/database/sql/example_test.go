// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

var db *sql.DB

func ExampleDB_QueryContext() {
	age := 27
	rows, err := db.QueryContext(context.Background(), "SELECT name FROM users WHERE age=?", age)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s are %d years old", strings.Join(names, ", "), age)
}

func ExampleDB_QueryRowContext() {
	id := 123
	var username string
	var created time.Time
	err := db.QueryRowContext(context.Background(), "SELECT username, created_at FROM users WHERE id=?", id).Scan(&username, &created)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No user with id %d", id)
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Username is %s, account created on %s\n", username, created)
	}
}

func ExampleDB_ExecContext() {
	ctx := context.Background()
	id := 47
	result, err := db.ExecContext(ctx, "UPDATE balances SET balance = balance + 10 WHERE user_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rows != 1 {
		panic(err)
	}
}

func ExampleDB_Query_multipleResultSets() {
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
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id   int64
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("id %d name is %s\n", id, name)
	}
	if !rows.NextResultSet() {
		log.Fatal("expected more result sets", rows.Err())
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
			log.Fatal(err)
		}
		fmt.Printf("id %d has role %s\n", id, roleMap[role])
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleDB_PingContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
}

func ExampleDB_Stats() {
	stats := db.Stats()
	fmt.Println(stats.OpenConnections)
}

func ExampleConn_BeginTx() {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Fatal(err)
	}
	id := 37
	_, execErr := tx.Exec(`UPDATE users SET status = 'paid' WHERE id = ?`, id)
	if execErr != nil {
		log.Fatal(execErr)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func ExampleConn_ExecContext() {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	id := 41
	result, err := conn.ExecContext(ctx, `UPDATE balances SET balance = balance + 10 WHERE user_id = ?`, id)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rows != 1 {
		panic(err)
	}
}

func ExampleTx_ExecContext() {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Fatal(err)
	}
	id := 37
	_, execErr := tx.ExecContext(ctx, "UPDATE users SET status = 'paid' WHERE id = ?", id)
	if execErr != nil {
		log.Fatal(execErr)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func ExampleTx_Rollback() {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Fatal(err)
	}
	id := 53
	_, err = tx.ExecContext(ctx, "UPDATE drivers SET status = 'assigned' WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.ExecContext(ctx, "UPDATE pickups SET driver_id = $1", id)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Could not roll back: %v\n", rollbackErr)
		}
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func ExampleStmt() {
	// In normal use, create one Stmt when your process starts
	stmt, err := db.PrepareContext(context.Background(), "SELECT username FROM users WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	// Then reuse it each time you need to issue the query.
	id := 43
	var username string
	err = stmt.QueryRowContext(context.Background(), id).Scan(&username)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No user with that ID.")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Username is %s\n", username)
	}
}

func ExampleStmt_QueryRowContext() {
	// In normal use, create one Stmt when your process starts
	stmt, err := db.PrepareContext(context.Background(), "SELECT username FROM users WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	// Then reuse it each time you need to issue the query.
	id := 43
	var username string
	err = stmt.QueryRowContext(context.Background(), id).Scan(&username)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("No user with that ID.")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Username is %s\n", username)
	}
}

func ExampleRows() {
	age := 27
	rows, err := db.QueryContext(context.Background(), "SELECT name FROM users WHERE age=?", age)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s are %d years old", strings.Join(names, ", "), age)
}
