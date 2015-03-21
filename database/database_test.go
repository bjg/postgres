package database

import (
	"log"
	"strings"
	"testing"
	"time"

	"fmt"

	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	ID      int64     `json:"id" db:"id" ddl:"bigserial PRIMARY KEY"`
	Created time.Time `json:"created_at" db:"created_at" ddl:"timestamp with time zone"`
	Updated time.Time `json:"updated_at" db:"updated_at" ddl:"timestamp with time zone"`
	Name    string    `json:"name" db:"name" ddl:"character varying(255)"`
	Email   string    `json:"email" db:"email" ddl:"character varying(255)"`
	Enabled bool      `json:"enabled" db:"enabled" ddl:"boolean DEFAULT false"`
}

func (m TestModel) GetInstance() interface{} {
	return &TestModel{}
}

func run(test func()) {
	MustCreateTable(TestModel{})
	test()
	DropTable(TestModel{})
}

func insertOne(name, email string) *TestModel {
	m := TestModel{Name: name, Email: email, Enabled: true}
	mc, err := Create(m)
	if err != nil {
		log.Fatal(err)
	}
	return mc.(*TestModel)
}

func TestCreate(t *testing.T) {
	run(func() {
		m := insertOne("John", "john@example.com")
		assert.NotEqual(t, 0, m.ID)
		assert.Equal(t, "John", m.Name)
		assert.Equal(t, "john@example.com", m.Email)
		assert.Equal(t, true, m.Enabled)
	})
}

func TestFindOne(t *testing.T) {
	run(func() {
		insertOne("John", "john@example.com")
		m := TestModel{}
		err := One(&m, `SELECT * FROM testmodels WHERE name = 'John' LIMIT 1`)
		if err != nil {
			t.Error(err)
		}
		assert.NotEqual(t, 0, m.ID)
		assert.Equal(t, "John", m.Name)
	})
}

func TestFindAll(t *testing.T) {
	run(func() {
		examples := []string{"Tom", "Dick", "Harry"}
		for _, name := range examples {
			insertOne(name, fmt.Sprintf("%s@example.com", strings.ToLower(name)))
		}
		ms := []TestModel{}
		err := Find(&ms, `SELECT * FROM testmodels`)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, len(examples), len(ms))
		for i, name := range examples {
			assert.Equal(t, name, ms[i].Name)
		}
	})
}

func TestSave(t *testing.T) {
	run(func() {
		m := insertOne("John", "john@example.com")
		m.Name = "Jane"
		if _, err := Save(*m); err != nil {
			t.Error(err)
		}
		s := TestModel{}
		err := One(&s, `SELECT * FROM testmodels WHERE id = $1 LIMIT 1`, m.ID)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, m.ID, s.ID)
		assert.Equal(t, "Jane", s.Name)
	})
}

func TestRemove(t *testing.T) {
	run(func() {
		m := insertOne("John", "john@example.com")
		if err := Remove(*m); err != nil {
			t.Error(err)
		}
		d := TestModel{}
		err := One(&d, `SELECT * FROM testmodels WHERE id = $1 LIMIT 1`, m.ID)
		assert.Error(t, err)
	})
}
