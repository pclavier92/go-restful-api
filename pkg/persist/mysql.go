package persist

import (
	"database/sql"
	"fmt"

	"github.com/pclavier92/go-restful-api/config"
	"github.com/pclavier92/go-restful-api/pkg/logs"

	// just importing mysql driver for the side effects :)
	_ "github.com/go-sql-driver/mysql"
)

// Querier is anything that can persist things and can tell us if there is something already there
type Querier interface {
	Query(q string, args ...interface{}) (*Rows, error)
	QueryRow(q string, arg ...interface{}) *Row
	Begin() (*Tx, error)
	Exec(q string, args ...interface{}) (Result, error)
	RowExists(where string, conditions string, args ...interface{}) bool
}

// Conn holds a connection to the database
type Conn struct {
	sql *sql.DB
	Log logs.Printer
}

// NewWithCustom lets you pass a custom *sql.DB to create a connection
func NewWithCustom(sql *sql.DB, logger logs.Printer) *Conn {
	return &Conn{sql, logger}
}

// New returns a new connectoin to the database with the defauly mysql connector
func New(cfg config.H, logger logs.Printer) (*Conn, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBName)
	mysql, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := mysql.Ping(); err != nil {
		return nil, err
	}
	mysql.SetMaxOpenConns(5)
	mysql.SetMaxIdleConns(5)
	return &Conn{mysql, logger}, nil
}

// NullString is a string which may be null on the database
type NullString struct {
	sql.NullString
}

// Stmt is a SQL statement
type Stmt struct {
	*sql.Stmt
}

// Tx represents a transaction in the DB
type Tx struct {
	t *sql.Tx
	l logs.Printer
}

// Prepare will prepare a query
func (t *Tx) Prepare(q string) (*Stmt, error) {
	s, e := t.t.Prepare(q)
	return &Stmt{s}, e
}

// Commit the transaction
func (t *Tx) Commit() error {
	return t.t.Commit()
}

// Exec something on the transaction
func (t *Tx) Exec(q string, args ...interface{}) (Result, error) {
	r, err := t.t.Exec(q, args...)
	return Result{r}, err
}

// Rollback the transaction. Takes the previous error so as to log both
// in case something goes wrong
func (t *Tx) Rollback(original error) {
	msg := ""
	if original != nil {
		msg = original.Error()
	}
	if err := t.t.Rollback(); err != nil {
		t.l.Info("There was a problem rollbacking!", logs.I{
			"error":    err.Error(),
			"original": msg,
		},
		)
	}
}

// Result is a result from the DB
type Result struct {
	sql.Result
}

// Rows is a collection of rows from the DB
type Rows struct {
	r *sql.Rows
	l logs.Printer
}

// Close will avoid further quering about the rows
func (r *Rows) Close() {
	if err := r.r.Close(); err != nil {
		r.l.Info("Error closing rows", logs.I{"err": err.Error()})
	}
	if err := r.r.Err(); err != nil {
		r.l.Info("Error when traversing rows", logs.I{"err": err.Error()})
	}
}

// Next will iterate through the next row
func (r *Rows) Next() bool {
	nxt := r.r.Next()
	if !nxt {
		r.Close()
	}
	return nxt
}

// Err will give you any error which may have happened
// while using Next() on the rows
func (r *Rows) Err() error {
	return r.r.Err()
}

// UnsafeScan will scan a row, but will NOT close them if an error ocurred
func (r *Rows) UnsafeScan(dest ...interface{}) error {
	return r.r.Scan(dest...)
}

// Scan will write the row values into the provided go variables
func (r *Rows) Scan(dest ...interface{}) error {
	err := r.r.Scan(dest...)
	if err != nil {
		r.Close()
	}
	return err
}

// Row represents a single row from the DB
type Row struct {
	*sql.Row
}

// Query the DB.
func (c *Conn) Query(q string, args ...interface{}) (*Rows, error) {
	rws, err := c.sql.Query(q, args...)
	return &Rows{rws, c.Log}, err
}

// QueryRow queries the DB about something which has to return ONE row.
func (c *Conn) QueryRow(q string, args ...interface{}) *Row {
	return &Row{c.sql.QueryRow(q, args...)}
}

// Begin a transaction.
func (c *Conn) Begin() (*Tx, error) {
	tx, err := c.sql.Begin()
	return &Tx{tx, c.Log}, err
}

// Exec will run the query inmediatly and return a result.
func (c *Conn) Exec(q string, args ...interface{}) (Result, error) {
	r, err := c.sql.Exec(q, args...)
	return Result{r}, err
}

// RowExists will tell you if the specified row exists in the "where" table, matching the condition string.
func (c *Conn) RowExists(where string, condition string, args ...interface{}) bool {
	var exists bool
	query := fmt.Sprintf("SELECT exists (SELECT 1 FROM %s WHERE %s)", where, condition)
	err := c.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return true
	}
	return exists
}
