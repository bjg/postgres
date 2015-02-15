package main

import (
	"encoding/json"
	"log"
	"os"

	"fmt"

	"flaconline.info/server/lib/database"
	"flaconline.info/server/lib/database/util"
	"flaconline.info/server/model"
)

var (
	categoryIds  = make(map[int64]int64)
	referrerIds  = make(map[int64]int64)
	referableIds = make(map[int64]int64)
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Missing one or more JSON file paths")
	}
	slurpCategories(loadData(os.Args[1]))
	slurpSections(loadData(os.Args[2]))
}

func loadData(fileName string) util.JsonData {
	fd, err := os.Open(fileName)
	defer fd.Close()
	if err != nil {
		log.Fatal(err)
	}
	data := util.JsonData{}
	if err = json.NewDecoder(fd).Decode(&data); err != nil {
		log.Fatal(err)
	}
	return data
}

func prologue(model interface{}) {
	if database.Exists(model) {
		log.Fatal(fmt.Sprintf("Table %s already exists", util.TableName(model)))
	}
	database.MustCreateTable(model)
}

func slurpCategories(data util.JsonData) {
	prologue(model.Category{})
	for _, m := range util.ImportJson(model.Category{}, data) {
		c := m.(model.Category)
		oldId := c.ID
		r := &model.Category{}
		err := database.Create(c).One(r)
		if err != nil {
			log.Fatal(err)
		}
		categoryIds[oldId] = r.ID
	}
}

func slurpSections(data util.JsonData) {
	prologue(model.Section{})
	for _, m := range util.ImportJson(model.Section{}, data) {
		s := m.(model.Section)
		oldId := s.ID
		s.CategoryID = categoryIds[s.CategoryID]
		r := &model.Section{}
		err := database.Create(s).One(r)
		if err != nil {
			log.Fatal(err)
		}
		referrerIds[oldId] = r.ID
	}
}
