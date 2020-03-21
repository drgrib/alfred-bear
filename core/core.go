package core

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/drgrib/alfred"
	"golang.org/x/text/unicode/norm"

	"github.com/drgrib/alfred-bear/db"
)

func getUniqueTagString(tagString string) string {
	if tagString == "" {
		return ""
	}
	tags := strings.Split(tagString, ",")
	uniqueTags := []string{}
	for _, t := range tags {
		isPrefix := false
		for _, other := range tags {
			if t != other && strings.HasPrefix(other, t) {
				isPrefix = true
				break
			}
		}
		if !isPrefix {
			uniqueTags = append(uniqueTags, t)
		}
	}
	sort.Strings(uniqueTags)
	return "#" + strings.Join(uniqueTags, " #")
}

func RowToItem(row map[string]string) alfred.Item {
	return alfred.Item{
		Title:    row[db.TitleKey],
		Subtitle: getUniqueTagString(row[db.TagsKey]),
		Arg:      row[db.NoteIDKey],
		Valid:    alfred.Bool(true),
	}
}

func AddNoteRowsToAlfred(rows []map[string]string) {
	for _, row := range rows {
		alfred.Add(RowToItem(row))
	}
}

type Query struct {
	Tokens, Tags          []string
	LastToken, WordString string
}

func ParseQuery(arg string) Query {
	query := Query{}
	query.Tokens = strings.Split(norm.NFC.String(arg), " ")
	query.Tags = []string{}
	words := []string{}
	for _, e := range query.Tokens {
		switch {
		case e == "":
		case strings.HasPrefix(e, "#"):
			query.Tags = append(query.Tags, e)
		default:
			words = append(words, e)
		}
	}
	query.LastToken = query.Tokens[len(query.Tokens)-1]
	query.WordString = strings.Join(words, " ")
	return query
}

func AutocompleteTags(litedb db.LiteDB, query Query) (bool, error) {
	if strings.HasPrefix(query.LastToken, "#") {
		rows, err := litedb.Query(fmt.Sprintf(db.TAGS_BY_TITLE, query.LastToken[1:]))
		if err != nil {
			return false, err
		}

		for _, row := range rows {
			tag := "#" + row[db.TitleKey]
			autocomplete := strings.Join(query.Tokens[:len(query.Tokens)-1], " ") + " " + tag + " "
			alfred.Add(alfred.Item{
				Title:        tag,
				Autocomplete: strings.TrimLeft(autocomplete, " "),
				Valid:        alfred.Bool(false),
				UID:          tag,
			})
		}
		return true, nil
	}
	return false, nil
}

func escape(s string) string {
	return strings.Replace(s, "'", "''", -1)
}

func GetSearchRows(litedb db.LiteDB, query Query) ([]map[string]string, error) {
	escapedWordString := escape(query.WordString)
	switch {
	case query.WordString == "" && len(query.Tags) == 0 && query.LastToken == "":
		rows, err := litedb.Query(db.RECENT_NOTES)
		if err != nil {
			return nil, err
		}
		return rows, nil

	case len(query.Tags) != 0:
		tagConditions := []string{}
		for _, t := range query.Tags {
			c := fmt.Sprintf("lower(tag.ZTITLE) = lower('%s')", t[1:])
			tagConditions = append(tagConditions, c)
		}
		tagConjunction := strings.Join(tagConditions, " OR ")
		rows, err := litedb.Query(fmt.Sprintf(db.NOTES_BY_TAGS_AND_QUERY, tagConjunction, escapedWordString, escapedWordString, len(query.Tags), escapedWordString))
		if err != nil {
			return nil, err
		}
		return rows, nil

	default:
		rows, err := litedb.Query(fmt.Sprintf(db.NOTES_BY_QUERY, escapedWordString, escapedWordString, escapedWordString))
		if err != nil {
			return nil, err
		}
		return rows, nil
	}
}

func GetCreateItem(query Query) (*alfred.Item, error) {
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
		return nil, err
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
	return &item, nil
}
