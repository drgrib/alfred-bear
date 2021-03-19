package core

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/drgrib/alfred"
	"github.com/pkg/errors"
	"golang.org/x/text/unicode/norm"

	"github.com/drgrib/alfred-bear/db"
)

var special = []string{
	"@tagged",
	"@untagged",
	"@today",
	"@yesterday",
	"@lastXdays",
	"@images",
	"@files",
	"@attachments",
	"@task",
	"@todo",
	"@done",
	"@code",
	"@title",
	"@locked",
	"@date(",
	"@cdate(",
}

const argSplit = "|"

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

func RowToItem(row db.Note, query Query) alfred.Item {
	searchCallbackString := getSearchCallbackString(query)
	return alfred.Item{
		Title:    row[db.TitleKey],
		Subtitle: getUniqueTagString(row[db.TagsKey]),
		Arg: strings.Join([]string{
			row[db.NoteIDKey],
			searchCallbackString,
		},
			argSplit,
		),
		Valid: alfred.Bool(true),
	}
}

func AddNoteRowsToAlfred(rows []db.Note, query Query) {
	for _, row := range rows {
		item := RowToItem(row, query)
		alfred.Add(item)
	}
}

type Query struct {
	Tokens     []string
	Tags       []string
	LastToken  string
	WordString string
}

func (query Query) String() string {
	return strings.Join(query.Tokens, " ")
}

var spaces = regexp.MustCompile(`\s+`) //nolint:gochecknoglobals

func ParseQuery(arg string) Query {
	query := Query{Tokens: spaces.Split(norm.NFC.String(arg), -1)}

	query.Tags = make([]string, 0, len(query.Tokens))
	words := make([]string, 0, len(query.Tokens))

	for _, e := range query.Tokens {
		switch {
		case strings.HasPrefix(e, "#"):
			query.Tags = append(query.Tags, e)
		default:
			words = append(words, e)
		}
	}

	query.LastToken = query.Tokens[len(query.Tokens)-1]
	query.WordString = strings.TrimSpace(strings.Join(words, " "))

	return query
}

func Autocomplete(litedb db.LiteDB, query Query) (bool, error) {
	autocompleted, err := AutocompleteTags(litedb, query)
	if err != nil {
		return false, err
	}
	if autocompleted {
		return autocompleted, nil
	}

	return AutocompleteSpecial(litedb, query)
}

func AutocompleteSpecial(litedb db.LiteDB, query Query) (bool, error) {
	if strings.HasPrefix(query.LastToken, "@") {
		for _, s := range special {
			if strings.HasPrefix(s, query.LastToken) {
				autocomplete := strings.Join(query.Tokens[:len(query.Tokens)-1], " ") + " " + s + " "
				alfred.Add(alfred.Item{
					Title:        s,
					Autocomplete: strings.TrimLeft(autocomplete, " "),
					Valid:        alfred.Bool(false),
					UID:          s,
				})
			}
		}
		return true, nil
	}

	if strings.HasPrefix(query.LastToken, "-@") {
		for _, s := range special {
			if strings.HasPrefix(s, query.LastToken[1:]) {
				s = "-" + s
				autocomplete := strings.Join(query.Tokens[:len(query.Tokens)-1], " ") + " " + s + " "
				alfred.Add(alfred.Item{
					Title:        s,
					Autocomplete: strings.TrimLeft(autocomplete, " "),
					Valid:        alfred.Bool(false),
					UID:          s,
				})
			}
		}
		return true, nil
	}

	return false, nil
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

func GetSearchRows(litedb db.LiteDB, query Query) ([]db.Note, error) {
	switch {
	case query.WordString == "" && len(query.Tags) == 0 && query.LastToken == "":
		rows, err := litedb.Query(db.RECENT_NOTES)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return rows, nil

	case len(query.Tags) != 0:
		rows, err := litedb.QueryNotesByTextAndTags(query.WordString, query.Tags)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return rows, nil

	default:
		rows, err := litedb.QueryNotesByText(query.WordString)
		if err != nil {
			return nil, errors.WithStack(err)
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

	title := fmt.Sprintf("Create %q", query.WordString)
	if strings.Contains(title, `\"`) {
		title = fmt.Sprintf("Create '%s'", query.WordString)
	}
	item := alfred.Item{
		Title: title,
		Arg:   callbackString,
		Valid: alfred.Bool(true),
	}
	if len(query.Tags) != 0 {
		item.Subtitle = strings.Join(query.Tags, " ")
	}
	return &item, nil
}

func getSearchCallbackString(query Query) string {
	callback := []string{}

	if query.WordString != "" {
		callback = append(callback, "term="+url.PathEscape(query.WordString))
	}

	if len(query.Tags) != 0 {
		callback = append(callback, "tag="+query.Tags[0][1:])
	}

	return strings.Join(callback, "&")
}

func GetAppSearchItem(query Query) (*alfred.Item, error) {
	title := "Search in Bear App"
	if query.WordString != "" {
		title = fmt.Sprintf("Search %#v in Bear App", query.WordString)
	}

	callbackString := getSearchCallbackString(query)

	item := alfred.Item{
		Title: title,
		Arg:   callbackString,
		Valid: alfred.Bool(true),
	}
	if len(query.Tags) != 0 {
		item.Subtitle = query.Tags[0]
	}
	return &item, nil
}
