package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/drgrib/alfred"

	"github.com/drgrib/alfred-bear/core"
	"github.com/drgrib/alfred-bear/db"
)

func main() {
	query := core.ParseQuery(os.Args[1])

	litedb, err := db.NewBearDB()
	if err != nil {
		panic(err)
	}

	autocompleted, err := core.AutocompleteTags(litedb, query)
	if err != nil {
		panic(err)
	}

	if !autocompleted {
		callback := []string{}
		if query.WordString != "" {
			callback = append(callback, "title="+url.PathEscape(query.WordString))
		}
		if len(query.Tags) != 0 {
			bareTags := []string{}
			for _, t := range query.Tags {
				bareTags = append(bareTags, url.PathEscape(t[1:]))
			}
			callback = append(callback, "tags="+strings.Join(bareTags, ","))
		}

		clipString, err := clipboard.ReadAll()
		if err != nil {
			panic(err)
		}
		if clipString != "" {
			callback = append(callback, "text="+url.PathEscape(clipString))
		}
		callbackString := strings.Join(callback, "&")

		item := alfred.Item{
			Title: fmt.Sprintf("Create %#v", query.WordString),
			Arg:   callbackString,
			Valid: alfred.Bool(true),
		}
		if len(query.Tags) != 0 {
			item.Subtitle = strings.Join(query.Tags, " ")
		}
		alfred.Add(item)
	}

	alfred.Run()
}
