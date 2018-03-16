// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

var (
	ctx context.Context
	db  *sql.DB
)

func ExampleDB_QueryContext() error {
	age := 27
	rows, err := db.QueryContext(ctx, "SELECT name FROM users WHERE age=?", age)
	if err != nil {
		return err
	}
	defer rows.Close()
	names := make([]string, 0)

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			return err
		}
		names = append(names, name)
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
	fmt.Printf("%s are %d years old", strings.Join(names, ", "), age)
	return nil
}

func ExampleDB_QueryRowContext() error {
	id := 123
	var username string
	var created time.Time
	err := db.QueryRowContext(ctx, "SELECT username, created_at FROM users WHERE id=?", id).Scan(&username, &created)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("no user with id %d\n", id)
		return nil
	case err != nil:
		fmt.Printf("query error: %v\n", err)
		return err
	default:
		fmt.Printf("username is %q, account created on %s\n", username, created)
		return nil
	}
}

func ExampleDB_ExecContext() error {
	id := 47
	result, err := db.ExecContext(ctx, "UPDATE balances SET balance = balance + 10 WHERE user_id = ?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}
	return nil
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

func ExampleConn_BeginTx() error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	id := 37
	_, execErr := tx.Exec(`UPDATE users SET status = ? WHERE id = ?`, "paid", id)
	if execErr != nil {
		_ = tx.Rollback()
		return execErr
	}
	return tx.Commit()
}

func ExampleConn_ExecContext() error {
	// A *DB is a pool of connections. Call Conn to reserve a connection for
	// exclusive use.
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close() // Return the connection to the pool.
	id := 41
	result, err := conn.ExecContext(ctx, `UPDATE balances SET balance = balance + 10 WHERE user_id = ?;`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected single row affected, got %d rows affected", rows)
	}
	return nil
}

func ExampleTx_ExecContext() error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	id := 37
	_, execErr := tx.ExecContext(ctx, "UPDATE users SET status = ? WHERE id = ?", "paid", id)
	if execErr != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("unable to rollback: %v\n", rollbackErr)
		}
		return execErr
	}
	return tx.Commit()
}

func ExampleTx_Rollback() error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	id := 53
	_, err = tx.ExecContext(ctx, "UPDATE drivers SET status = ? WHERE id = ?;", "assigned", id)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("update drivers: unable to rollback: %v", rollbackErr)
		}
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE pickups SET driver_id = $1;", id)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("update pickups: unable to back: %v", rollbackErr)
		}
		return err
	}
	return tx.Commit()
}

func ExampleStmt() error {
	// In normal use, create one Stmt when your process starts.
	stmt, err := db.PrepareContext(ctx, "SELECT username FROM users WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Then reuse it each time you need to issue the query.
	id := 43
	var username string
	err = stmt.QueryRowContext(ctx, id).Scan(&username)
	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("no user with id %d", id)
	case err != nil:
		return err
	default:
		fmt.Printf("Username is %s\n", username)
		return nil
	}
}

func ExampleStmt_QueryRowContext() error {
	// In normal use, create one Stmt when your process starts.
	stmt, err := db.PrepareContext(ctx, "SELECT username FROM users WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Then reuse it each time you need to issue the query.
	id := 43
	var username string
	err = stmt.QueryRowContext(ctx, id).Scan(&username)
	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("no user with id %d", id)
	case err != nil:
		return err
	default:
		fmt.Printf("Username is %s\n", username)
		return nil
	}
}

func ExampleRows() error {
	age := 27
	rows, err := db.QueryContext(ctx, "SELECT name FROM users WHERE age=?", age)
	if err != nil {
		return err
	}
	defer rows.Close()

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		names = append(names, name)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return err
	}
	fmt.Printf("%s are %d years old", strings.Join(names, ", "), age)
	return nil
}
