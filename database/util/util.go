package util

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/bjg/postgres/database/types"
	"github.com/gedex/inflector"
	_ "github.com/lib/pq"
)

type (
	columnInfo map[string]int

	JsonFieldsValuesFormat struct {
		Fields []string   `json:"fields"`
		Values [][]string `json:"values"`
	}
)

var (
	tableNames       = make(map[interface{}]string)
	insertStatements = make(map[string]string)
	updateStatements = make(map[string]string)
)

func MakeCreate(model interface{}) string {
	t := reflect.TypeOf(model)
	n := t.NumField()
	var attrs bytes.Buffer
	for i := 0; i < n; i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")
		attrs.WriteString(tag)
		attrs.WriteString(" ")
		ddl := f.Tag.Get("ddl")
		attrs.WriteString(ddl)
		if i+1 < n {
			attrs.WriteString(", ")
		}
	}
	return attrs.String()
}

func MakeInsert(model interface{}) (string, []interface{}) {
	name := TableName(model)
	if _, ok := insertStatements[name]; !ok {
		var (
			attrs, bindvars bytes.Buffer
			cnt             int
		)
		const sep = ", "
		t := reflect.TypeOf(model)
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			tag := f.Tag.Get("db")
			if tag != "id" {
				attrs.WriteString(tag)
				if matches, _ := regexp.MatchString("(created|updated)_at", tag); matches {
					bindvars.WriteString("NOW()")
				} else {
					cnt++
					bindvars.WriteString(fmt.Sprintf("$%d", cnt))
				}
				if i+1 < n {
					attrs.WriteString(sep)
					bindvars.WriteString(sep)
				}
			}
		}
		stmt := fmt.Sprintf(`INSERT INTO %s (%v) VALUES (%v) RETURNING *`,
			name, attrs.String(), bindvars.String())
		insertStatements[name] = stmt
	}
	return insertStatements[name], getValues(model, "^id$|(created|updated)_at")
}

func MakeUpdate(model interface{}) (string, []interface{}) {
	name := TableName(model)
	if _, ok := updateStatements[name]; !ok {
		var (
			attrs bytes.Buffer
			cnt   int
		)
		const sep = ", "
		t := reflect.TypeOf(model)
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			tag := f.Tag.Get("db")
			if tag != "id" && tag != "" && tag != "-" {
				attrs.WriteString(tag)
				attrs.WriteString(" = ")
				if matches, _ := regexp.MatchString("_at", tag); matches {
					attrs.WriteString("NOW()")
				} else {
					cnt++
					attrs.WriteString(fmt.Sprintf("$%d", cnt))
				}
				if i+1 < n {
					attrs.WriteString(sep)
				}
			}
		}
		stmt := fmt.Sprintf(`UPDATE %s SET %v WHERE id = %v RETURNING *`,
			name, attrs.String(), fmt.Sprintf("$%d", cnt+1))
		updateStatements[name] = stmt
	}
	return updateStatements[name], append(getValues(model, "^id$|(created|updated)_at"), GetValueOfField(model, "ID"))
}

func TableName(model interface{}) string {
	if _, ok := tableNames[model]; !ok {
		name := reflect.TypeOf(model).String()
		name = name[strings.LastIndex(name, ".")+1:]
		tableNames[model] = strings.ToLower(inflector.Pluralize(name))
	}
	return tableNames[model]
}

func GetValueOfField(model interface{}, fieldName string) interface{} {
	return reflect.ValueOf(model).FieldByName(fieldName).Interface()
}

func GetNamedTagForField(model interface{}, fieldName, tagName string) string {
	t := reflect.TypeOf(model)
	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.Name == fieldName {
			return f.Tag.Get(tagName)
		}
	}
	return ""
}

func getValues(model interface{}, exclude string) []interface{} {
	t := reflect.TypeOf(model)
	n := t.NumField()
	v := reflect.ValueOf(model)
	var values []interface{}
	for i := 0; i < n; i++ {
		if matches, _ := regexp.MatchString(exclude, t.Field(i).Tag.Get("db")); !matches {
			f := v.Field(i).Interface()
			// XXX Handle bespoke type conversions
			switch f.(type) {
			case types.Timestamp:
				f = f.(types.Timestamp).String()
			case types.WallClock:
				f = f.(types.WallClock).String()
			case types.Array:
				f = f.(types.Array).String()
			}
			values = append(values, f)
		}
	}
	return values
}

func ImportJsonFieldsValues(model interface{}, data JsonFieldsValuesFormat, unmarshal func([]byte) interface{}) []interface{} {
	names := fieldNamesByTag(model, "json")
	all := make([]interface{}, len(data.Values))
	for rowNum, rowVal := range data.Values {
		var buf bytes.Buffer
		cnt := 0
		buf.WriteString("{")
		for fieldNum, fieldVal := range rowVal {
			fieldName := data.Fields[fieldNum]
			if _, ok := names[fieldName]; ok {
				if cnt != 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(fmt.Sprintf(`"%s": "%s"`, fieldName, fieldVal))
				cnt++
			}
		}
		buf.WriteString("}")
		all[rowNum] = unmarshal(buf.Bytes())
	}
	return all
}

func fieldNamesByTag(model interface{}, tagName string) map[string]string {
	t := reflect.TypeOf(model)
	n := t.NumField()
	names := make(map[string]string)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		tag := f.Tag.Get(tagName)
		if tag != "" && tag != "-" {
			s := regexp.MustCompile(",").Split(tag, -1)
			names[s[0]] = f.Name
		}
	}
	return names
}
