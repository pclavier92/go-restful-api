package sqltest

import (
	"database/sql/driver"

	"github.com/pclavier92/go-restful-api/pkg/errors"
	"github.com/pclavier92/go-restful-api/pkg/logs"
	"github.com/pclavier92/go-restful-api/pkg/persist"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// Mock allows you to mock SQL queries and specify which rows they should return!
type Mock struct {
	e errors.Structer
	sqlmock.Sqlmock
}

// Rows is a collection of results from an SQL query
type Rows [][]interface{}

// Row is a single row from an SQL query
type Row []interface{}

// Args are the arguments of an SQL query
type Args []interface{}

func (m Mock) toValues(is [][]interface{}) ([][]driver.Value, error) {
	e := m.e.Fn("toValues")
	var values [][]driver.Value
	for _, v := range is {
		var vv []driver.Value
		for _, arg := range v {
			if arg == nil {
				vv = append(vv, arg)
				continue
			}
			d, ok := arg.(driver.Value)
			if !ok {
				return values, e.New("cant convert to driver.Value")
			}
			vv = append(vv, d)
		}
		values = append(values, vv)
	}
	return values, nil
}

// ExpectSelect will assert a select is comming. Will only work with queries.
func (m Mock) ExpectSelect(cls []string, res Rows) error {
	rws := sqlmock.NewRows(cls)
	values, err := m.toValues([][]interface{}(res))
	if err != nil {
		return err
	}
	for _, v := range values {
		rws.AddRow(v...)
	}
	m.ExpectQuery("SELECT *").WillReturnRows(rws)
	return nil
}

// ExpectUpdate will assert an update is comming. Will only work with Exec.
func (m Mock) ExpectUpdate(to Args) error {
	values, err := m.toValues([][]interface{}{[]interface{}(to)})
	if err != nil {
		return err
	}
	var vs []driver.Value
	for _, v := range values {
		vs = append(vs, v...)
	}
	m.ExpectExec("UPDATE *").WithArgs(vs...).WillReturnResult(sqlmock.NewResult(1, 1))
	return nil
}

// ExpectPreparedExec will assert an insert is coming on a prepared statement.
func (m Mock) ExpectPreparedExec(s *sqlmock.ExpectedPrepare, to Args) error {
	values, err := m.toValues([][]interface{}{[]interface{}(to)})
	if err != nil {
		return err
	}
	var vs []driver.Value
	for _, v := range values {
		vs = append(vs, v...)
	}
	s.ExpectExec().WithArgs(vs...).WillReturnResult(sqlmock.NewResult(1, 1))
	return nil
}

// ExpectInsert will assert an insert is comming. Will only work with Exec.
func (m Mock) ExpectInsert(to Args) error {
	values, err := m.toValues([][]interface{}{[]interface{}(to)})
	if err != nil {
		return err
	}
	var vs []driver.Value
	for _, v := range values {
		vs = append(vs, v...)
	}
	m.ExpectExec("INSERT *").WithArgs(vs...).WillReturnResult(sqlmock.NewResult(1, 1))
	return nil
}

// Tester tests databases
type DB struct {
	*persist.Conn
	truer          []bool
	errer          bool
	rowExistsCount int
}

// New returns a new tester
func New(rowExistsResponses []bool, errer bool) (*DB, *Mock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	log, err := logs.New("")
	if err != nil {
		return nil, nil, err
	}
	e := errors.Pkg("persisttest", log)
	p := persist.NewWithCustom(db, log)
	return &DB{p, rowExistsResponses, errer, 0}, &Mock{e.Struct("Mock"), mock}, nil
}

func (db *DB) ExpectRowExists(res bool) {
	db.truer = append(db.truer, res)
}

// RowExists will tell you if a row exists according to the config of the tester
func (db *DB) RowExists(where string, conditions string, args ...interface{}) bool {
	r := db.truer[db.rowExistsCount]
	db.rowExistsCount++
	return r
}
