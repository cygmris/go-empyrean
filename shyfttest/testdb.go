package shyfttest

import (
	"os/exec"
	"time"

	"gopkg.in/khaiql/dbcleaner.v2"
	"gopkg.in/khaiql/dbcleaner.v2/engine"
)

//@SHYFT NOTE: Side effects from PG database therefore need to reset before running

// Cleaner - wrapper for testdb
var Cleaner = dbcleaner.New(dbcleaner.SetNumberOfRetry(10), dbcleaner.SetLockTimeout(5*time.Second))

const connStrTest = "user=postgres dbname=shyftdbtest password=docker sslmode=disable"

// PgTestDbSetup - reinitializes the pg database
func PgTestDbSetup() {

	cmdStr := "$GOPATH/src/github.com/ShyftNetwork/go-empyrean/shyftdb/postgres_setup_test/init_test_db.sh"
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	_, err := cmd.Output()
	PgRecreateTables()
	pg := engine.NewPostgresEngine(connStrTest)
	Cleaner.SetEngine(pg)
	Cleaner.Acquire("accounts")
	Cleaner.Acquire("blocks")
	Cleaner.Acquire("txs")
	Cleaner.Acquire("internaltxs")
	if err != nil {
		println(err.Error())
		return
	}
}

// PgTestTearDown - resets the pg test database
func PgTestTearDown() {
	Cleaner.Clean("accounts")
	Cleaner.Clean("blocks")
	Cleaner.Clean("txs")
	Cleaner.Clean("internaltxs")
}

// PgRecreateTables - recreates pg database tables
func PgRecreateTables() {
	cmdStr := "$GOPATH/src/github.com/ShyftNetwork/go-empyrean/shyftdb/postgres_setup_test/recreate_tables_test.sh"
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	_, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return
	}
}

// Cleans All the Tables In Tests
func TruncateTables() {
	Cleaner.Acquire("accounts")
	Cleaner.Acquire("blocks")
	Cleaner.Acquire("txs")
	Cleaner.Acquire("internaltxs")
}
