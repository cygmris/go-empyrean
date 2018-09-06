package shyftdb

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/ShyftNetwork/go-empyrean/core"
)

func TestDbCreationExistence(t *testing.T) {

	t.Run("Creates the PG DB if it Doesnt Exist", func(t *testing.T) {
		core.DeletePgDb(core.DbName())
		db, err := core.InitDB()
		if err != nil || err == sql.ErrNoRows {
			fmt.Println(err)
		}
		db.Close()
		_, err = core.DbExists(core.DbName())
		if err != nil || err == sql.ErrNoRows {
			t.Errorf("Error in Database Connection - DB doesn't Exist - %s", err)
		}
		core.DeletePgDb(core.DbName())
	})
	t.Run("Creates the Tables Required from the Migration Schema", func(t *testing.T) {
		db, err := core.InitDB()
		if err != nil || err == sql.ErrNoRows {
			fmt.Println(err)
		}
		db.Close()
		tableNameQuery := `SELECT tablename FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema='public';`
		db = core.Connect(core.ConnectionStr())
		defer db.Close()
		if err != nil || err == sql.ErrNoRows {
			t.Errorf("Error in Database Connection - %s", err)
		}
		rows, err := db.Query(tableNameQuery)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		var tablenames string
		var table string
		for rows.Next() {
			err = rows.Scan(&table)
			tablenames += table + ","
			if err != nil {
				panic(err)
			}
			fmt.Println(tablenames)
		}
		err = rows.Err()
		if err != nil {
			panic(err)
		}
	})
	t.Run("If the Database Doesnt Exist It Creates It", func(t *testing.T) {

	})
}
