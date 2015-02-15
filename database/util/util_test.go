package util

import (
	"testing"

	"flaconline.info/server/lib/database/types"
	"github.com/stretchr/testify/assert"
)

func TestTableName(t *testing.T) {
	type (
		Model  struct{}
		Person struct{}
	)
	for _, example := range []struct {
		model interface{}
		name  string
	}{
		{Model{}, "models"},
		{&Model{}, "models"},
		{&[]Model{}, "models"},
		{Person{}, "people"},
	} {
		assert.Equal(t, TableName(example.model), example.name)
	}
}

func TestMakeSelect(t *testing.T) {
	type Model struct{}
	for _, example := range []struct {
		cond []interface{}
		stmt string
	}{
		{[]interface{}{}, "SELECT to_json(models) FROM models"},
		{[]interface{}{"WHERE name = '$1'", "Bob"}, "SELECT to_json(models) FROM models WHERE name = '$1'"},
	} {
		assert.Equal(t, example.stmt, MakeSelect(Model{}, example.cond...))
	}
}

func TestMakeInsert(t *testing.T) {
	type User struct {
		ID      int64           `json:"id"`
		Name    string          `json:"name"`
		Email   string          `json:"email"`
		Created types.Timestamp `json:"created_at"`
	}
	stmt, values := MakeInsert(User{
		ID:    1,
		Name:  "John Smith",
		Email: "john@example.com",
	})
	assert.Contains(t, stmt, "INSERT INTO users (name, email, created_at) VALUES ($1, $2, NOW()) RETURNING *")
	assert.Len(t, values, 2)
}

func TestMakeUpdate(t *testing.T) {
	type User struct {
		ID      int64           `json:"id"`
		Name    string          `json:"name"`
		Email   string          `json:"email"`
		Created types.Timestamp `json:"created_at"`
	}
	u := User{
		ID:    1,
		Name:  "John Smith",
		Email: "john@example.com",
	}
	stmt, values := MakeUpdate(u)
	assert.Contains(t, stmt, "UPDATE users SET name = $1, email = $2, created_at = NOW() WHERE id = $3 RETURNING *")
	assert.Len(t, values, 3)
}

func TestMakeCreate(t *testing.T) {
	type User struct {
		ID      int64           `json:"id" ddl:"integer PRIMARY KEY NOT NULL DEFAULT nextval('%v_id_seq')"`
		Name    string          `json:"name" ddl:"character varying(255)"`
		Email   string          `json:"email" ddl:"character varying(255)"`
		Created types.Timestamp `json:"created_at" ddl:"timestamp with time zone"`
		Updated types.Timestamp `json:"updated_at" ddl:"timestamp with time zone"`
	}
	stmt := MakeCreate(User{})
	expected := `id integer PRIMARY KEY NOT NULL DEFAULT nextval('users_id_seq'), name character varying(255), email character varying(255), created_at timestamp with time zone, updated_at timestamp with time zone`
	assert.Equal(t, expected, stmt)
}
