package shyftdb

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/ShyftNetwork/go-empyrean/core"
)

func TestDbCreationExistence(t *testing.T) {
	core.DeletePgDb(core.DbName())
	db, err := core.InitDB()
	if err != nil || err == sql.ErrNoRows {
		fmt.Println(err)
	}
	t.Run("Creates the PG DB if it Doesnt Exist", func(t *testing.T) {
		_, err = core.DbExists(core.DbName())
		if err != nil || err == sql.ErrNoRows {
			t.Errorf("Error in Database Connection - DB doesn't Exist - %s", err)
		}
	})
	t.Run("Creates the Tables Required from the Migration Schema", func(t *testing.T) {
		db, err := core.InitDB()
		if err != nil || err == sql.ErrNoRows {
			fmt.Println(err)
		}
		tableNameQuery := `select table_name from information_schema.tables where table_schema = 'public' AND table_type = 'BASE TABLE' order by table_name ASC;`
		db = core.Connect(core.ShyftConnectStr())
		rows, err := db.Query(tableNameQuery)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		var tablenames string
		var table string
		notLast := rows.Next()
		for notLast {
			//... rows.Scan
			err = rows.Scan(&table)
			if err != nil {
				panic(err)
			}
			notLast = rows.Next()
			if notLast {
				tablenames += table + ", "
			} else {
				tablenames += table
			}
		}
		err = rows.Err()
		if err != nil {
			panic(err)
		}
		want := "accounts, blocks, internaltxs, txs"
		if tablenames != want {
			t.Errorf("Test Failed as wanted: %s  - got: %s", want, tablenames)
		}
	})
	t.Run("If the Database Doesnt Exist It Creates It", func(t *testing.T) {

	})
	core.DeletePgDb(core.DbName())
	db, err = core.InitDB()
	if err != nil || err == sql.ErrNoRows {
		fmt.Println(err)
	}
	db.Close()

}
