package database

import (
	"log"
	"strings"
	"testing"
	"time"

	"fmt"

	"flaconline.info/server/lib/database/types"
	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	ID      int64           `json:"id" ddl:"integer PRIMARY KEY NOT NULL DEFAULT nextval('%v_id_seq')"`
	Created types.Timestamp `json:"created_at" ddl:"timestamp with time zone"`
	Updated types.Timestamp `json:"updated_at" ddl:"timestamp with time zone"`
	Name    string          `json:"name" ddl:"character varying(255)"`
	Email   string          `json:"email" ddl:"character varying(255)"`
	Enabled bool            `json:"enabled" ddl:"boolean DEFAULT false"`
}

func (tm TestModel) GetInstance() interface{} {
	return &TestModel{}
}

func run(test func()) {
	DropTable(TestModel{})
	MustCreateTable(TestModel{})
	test()
}

func insertOne(name, email string) *TestModel {
	m := &TestModel{}
	err := Create(TestModel{Name: name, Email: email, Enabled: true}).One(m)
	if err != nil {
		log.Fatal(err)
	}
	return &m
}

func TestCreate(t *testing.T) {
	run(func() {
		m := insertOne("John", "john@example.com")
		assert.NotEqual(t, 0, m.ID)
		assert.Equal(t, "John", m.Name)
		assert.Equal(t, "john@example.com", m.Email)
		assert.Equal(t, true, m.Enabled)
		assert.Equal(t, time.Now().Second(), m.Created.Second())
	})
}

func TestFindOne(t *testing.T) {
	run(func() {
		insertOne("John", "john@example.com")
		m := &TestModel{}
		if err := Find(m, `WHERE name = 'John' LIMIT 1`).One(m); err != nil {
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
		ms := []*TestModel{}
		err := Find(TestModel{}).Each(TestModel{}, func(model interface{}) {
			ms = append(ms, model.(*TestModel))
		})
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, len(examples), len(ms))
		for i, name := range examples {
			assert.Equal(t, name, ms[i].Name)
		}
	})
}

func TestFetch(t *testing.T) {
	run(func() {
		id := insertOne("John", "john@example.com").ID
		m := &TestModel{}
		if err := Fetch(m, id); err != nil {
			t.Error(err)
		}
		assert.Equal(t, id, m.ID)
		assert.Equal(t, "John", m.Name)
	})
}

func TestSave(t *testing.T) {
	run(func() {
		m := insertOne("John", "john@example.com")
		m.Name = "Jane"
		s := &TestModel{}
		if err := Save(*m).One(s); err != nil {
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
		r := &TestModel{}
		err := Find(m, "WHERE id = $1 LIMIT 1", m.ID).One(r)
		assert.Error(t, err)
	})
}

/*
func TestDuplicate(t *testing.T) {
	run(func() {
		c, _ := insertOne("John", "john@example.com")
		_, err := c.Create()
		if err == nil {
			t.Error("Expected duplicate category to be rejected")
		}
	})
}

func TestNonDuplicate(t *testing.T) {
	run(func() {
		insertOne("John", "john@example.com")
		_, err := insertOne("Civil", "Civil Legal Issues", 0x111111, 0xeeeeee)
		if err != nil {
			t.Error("Expected non-duplicate category to be saved", "\n\n", err)
		}
	})
}

func TestUpdate(t *testing.T) {
	run(func() {
		c, _ := insertOne("John", "john@example.com")
		c.Background = 0x12345678
		u, err := c.Upsert()
		if err != nil {
			t.Error(err)
		}
		if u.Background != 0x12345678 {
			t.Error("Expected category attribute to be changed\n", *u)
		}
	})
}

*/
