package database

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

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
	spec := util.MakeCreate(model)
	db.MustExec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %v (%v)`, name, spec))
}

func DropTable(model interface{}) {
	db.MustExec(fmt.Sprintf(`DROP TABLE IF EXISTS %v RESTRICT`, util.TableName(model)))
}

func Truncate(model interface{}) error {
	_, err := db.Exec(fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY`, util.TableName(model)))
	return err
}

func Create(model Model) (interface{}, error) {
	stmt, values := util.MakeInsert(model)
	//log.Println(stmt, values)
	created := model.GetInstance()
	err := db.QueryRowx(stmt, values...).StructScan(created)
	return created, err
}

func Save(model Model) (interface{}, error) {
	stmt, values := util.MakeUpdate(model)
	updated := model.GetInstance()
	err := db.QueryRowx(stmt, values...).StructScan(updated)
	return updated, err
}

func Remove(model interface{}) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %v WHERE id = $1", util.TableName(model)),
		util.GetValueOfField(model, "ID"))
	return err
}

func One(model interface{}, query string, args ...interface{}) error {
	return db.Get(model, query, args...)
}

func Find(model interface{}, query string, args ...interface{}) error {
	return db.Select(model, query, args...)
}

func Exec(query string, args ...interface{}) error {
	_, err := db.Exec(query, args...)
	return err
}

func Prepare(query string) (*sqlx.Stmt, error) {
	return db.Preparex(query)
}

func Count(model interface{}, query string, args ...interface{}) (int, error) {
	var count int
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %v %v", util.TableName(model), query)
	err := db.QueryRowx(stmt, args...).Scan(&count)
	return count, err
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
