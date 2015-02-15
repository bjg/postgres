package database

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/bjg/postgres/database/types"
	"github.com/bjg/postgres/database/util"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

func init() {
	db = mustOpen()
}

func Exists(model interface{}) bool {
	var exists bool
	db.QueryRowx(fmt.Sprintf(`SELECT EXISTS(SELECT relname FROM pg_class 
		WHERE relname = '%s' AND relkind='r')`, util.TableName(model))).Scan(&exists)
	return exists
}

func MustNotExist(model interface{}) {
	if Exists(model) {
		log.Fatalf("Table %s still exists!\n", util.TableName(model))
	}
}

func MustCreateTable(model interface{}) {
	name := util.TableName(model)
	ddl := util.GetNamedTagForField(model, "id", "ddl")
	if match, _ := regexp.MatchString(ddl, "^uuid"); match {
		db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	}
	db.MustExec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %v (%v)`, name, util.MakeCreate(model)))
}

func DropTable(model interface{}) {
	db.MustExec(fmt.Sprintf(`DROP TABLE IF EXISTS %v RESTRICT`, util.TableName(model)))
}

func Truncate(model interface{}) error {
	_, err := db.Exec(fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY`, util.TableName(model)))
	return err
}

func Find(model interface{}, cond ...interface{}) *types.Result {
	stmt := util.MakeSelect(model, cond...)
	var (
		rows *sqlx.Rows
		err  error
	)
	if len(cond) == 0 {
		rows, err = db.Queryx(stmt)
	} else {
		rows, err = db.Queryx(stmt, cond[1:]...)
	}
	return readRows(rows, err)
}

func Create(model interface{}) *types.Result {
	stmt, values := util.MakeInsert(model)
	var enc string
	err := db.QueryRowx(stmt, values...).Scan(&enc)
	r := types.NewResult(err)
	return r.Update(enc)
}

func Save(model interface{}) *types.Result {
	stmt, values := util.MakeUpdate(model)
	var enc string
	err := db.QueryRowx(stmt, values...).Scan(&enc)
	r := types.NewResult(err)
	return r.Update(enc)
}

func Fetch(model interface{}, ID int64) error {
	return Find(model, "WHERE id = $1 LIMIT 1", ID).One(model)
}

func Remove(model interface{}) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %v WHERE id = $1", util.TableName(model)), util.GetValueOfField(model, "ID"))
	return err
}

func Query(stmt string, args ...interface{}) *types.Result {
	return readRows(db.Queryx(stmt, args...))
}

func mustOpen() *sqlx.DB {
	if db == nil {
		var err error
		uri := os.Getenv("DATABASE_URL")
		if uri == "" {
			log.Fatal("DATABASE_URL is not defined")
		}
		sslMode := "require"
		if strings.Contains(uri, "localhost") {
			sslMode = "disable"
		}
		db, err = sqlx.Open("postgres", fmt.Sprintf("%s?sslmode=%s", uri, sslMode))
		if err != nil {
			log.Fatal(err)
		}
	}
	return db
}

func readRows(rows *sqlx.Rows, err error) *types.Result {
	r := types.NewResult(err)
	for r.Error() == nil && rows.Next() {
		var enc string
		if rows.Scan(&enc) == nil {
			r.Update(enc)
		}
	}
	return r
}
