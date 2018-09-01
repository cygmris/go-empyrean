package core

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	// PG db connection adapter
	_ "github.com/lib/pq"
)

// @NOTE: SHYFT - could be refactored to add a test db environment
const (
	connStrTest    = "user=postgres dbname=shyftdbtest password=docker sslmode=disable"
	connStrDocker  = "user=postgres dbname=shyftdb host=pg password=docker sslmode=disable"
	connStrDefault = "user=postgres dbname=shyftdb host=localhost sslmode=disable"
)

var blockExplorerDb *sql.DB

// ConnStr String Made available for Testing & DbCleaner
var connStr = connectionStr()

// InitDB - initalizes a Postgresql Database for use by the Blockexplorer
func InitDB() (*sql.DB, error) {
	// To set the environment you can run the program with an ENV variable DBENV.
	// DBENV defaults to local for purposes of running the correct local
	// database connection parameters but will use docker connection parameters if DBENV=docker
	//
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("ERROR OPENING DB, NOT INITIALIZING")
		panic(err)
	}
	blockExplorerDb = db
	return blockExplorerDb, nil
}

func connectionStr() string {
	dbEnv := os.Getenv("DBENV")
	if flag.Lookup("test.v") == nil {
		switch dbEnv {
		default:
			return connStrDefault
		case "docker":
			return connStrDocker
		}
	} else {
		return connStrTest
	}
}

func DBConnection() (*sql.DB, error) {
	if blockExplorerDb == nil {
		_, err := InitDB()
		if err != nil {
			return nil, err
		}
	}
	return blockExplorerDb, nil
}

func ClearTables() {
	sqldb, err := DBConnection()
	tx, _ := sqldb.Begin()
	if err != nil {
		panic(err)
	}

	sqlStatementTx := `DELETE FROM txs`
	_, err = tx.Exec(sqlStatementTx)
	if err != nil {
		panic(err)
	}

	sqlStatementAcc := `DELETE FROM accounts`
	_, err = tx.Exec(sqlStatementAcc)
	if err != nil {
		panic(err)
	}

	sqlStatement := `DELETE FROM blocks`
	_, err = tx.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}
	tx.Commit()
}

// TruncateTables - Is primarily user to clear the pg database between unit tests
func TruncateTables() {
	sqldb, err := DBConnection()
	if err != nil {
		panic(err)
	}
	tx, _ := sqldb.Begin()
	sqlStatement := `TRUNCATE TABLE txs, accounts, blocks, internaltxs RESTART IDENTITY CASCADE;`
	_, err = tx.Exec(sqlStatement)
	tx.Commit()
	if err != nil {
		panic(err)
	}
}
