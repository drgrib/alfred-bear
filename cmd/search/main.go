package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/drgrib/alfred"
	_ "github.com/mattn/go-sqlite3"

	"github.com/drgrib/alfred-bear/comp"
)

const (
	DbPath = "~/Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data/database.sqlite"

	TitleKey  = "ZTITLE"
	TagsKey   = "group_concat(tag.ZTITLE)"
	NoteIDKey = "ZUNIQUEIDENTIFIER"

	RECENT_NOTES = `
SELECT DISTINCT
	note.ZUNIQUEIDENTIFIER, note.ZTITLE, group_concat(tag.ZTITLE)
FROM
	ZSFNOTE note
	INNER JOIN Z_7TAGS nTag ON note.Z_PK = nTag.Z_7NOTES
	INNER JOIN ZSFNOTETAG tag ON nTag.Z_14TAGS = tag.Z_PK
WHERE
	note.ZARCHIVED=0
	AND note.ZTRASHED=0
GROUP BY note.ZUNIQUEIDENTIFIER
ORDER BY
	note.ZMODIFICATIONDATE DESC
LIMIT 25
`

	NOTES_BY_QUERY = `
SELECT DISTINCT
	note.ZUNIQUEIDENTIFIER, note.ZTITLE, group_concat(tag.ZTITLE)
FROM
	ZSFNOTE note
	INNER JOIN Z_7TAGS nTag ON note.Z_PK = nTag.Z_7NOTES
	INNER JOIN ZSFNOTETAG tag ON nTag.Z_14TAGS = tag.Z_PK
WHERE
	note.ZARCHIVED=0
	AND note.ZTRASHED=0
	AND (
		lower(note.ZTITLE) LIKE lower('%%%s%%') OR
		lower(note.ZTEXT) LIKE lower('%%%s%%')
	)
GROUP BY note.ZUNIQUEIDENTIFIER
ORDER BY case when lower(note.ZTITLE) LIKE lower('%%%s%%') then 0 else 1 end, note.ZMODIFICATIONDATE DESC
LIMIT 25
`

	NOTES_BY_TAGS_AND_QUERY = `
SELECT DISTINCT
	note.ZUNIQUEIDENTIFIER, note.ZTITLE, group_concat(tag.ZTITLE)
FROM
	ZSFNOTE note
	INNER JOIN Z_7TAGS nTag ON note.Z_PK = nTag.Z_7NOTES
	INNER JOIN ZSFNOTETAG tag ON nTag.Z_14TAGS = tag.Z_PK
WHERE note.ZUNIQUEIDENTIFIER IN (
	SELECT
		note.ZUNIQUEIDENTIFIER
	FROM
		ZSFNOTE note
		INNER JOIN Z_7TAGS nTag ON note.Z_PK = nTag.Z_7NOTES
		INNER JOIN ZSFNOTETAG tag ON nTag.Z_14TAGS = tag.Z_PK
	WHERE
		note.ZARCHIVED=0
		AND note.ZTRASHED=0
		AND (%s)
		AND (
			lower(note.ZTITLE) LIKE lower('%%%s%%') OR
			lower(note.ZTEXT) LIKE lower('%%%s%%')
		)
	GROUP BY note.ZUNIQUEIDENTIFIER
	HAVING COUNT(*) >= %d
)
GROUP BY note.ZUNIQUEIDENTIFIER
ORDER BY case when lower(note.ZTITLE) LIKE lower('%%%s%%') then 0 else 1 end, note.ZMODIFICATIONDATE DESC
LIMIT 25
`

	TAGS_BY_TITLE = `
SELECT DISTINCT
	t.ZTITLE
FROM
	ZSFNOTE n
	INNER JOIN Z_7TAGS nt ON n.Z_PK = nt.Z_7NOTES
	INNER JOIN ZSFNOTETAG t ON nt.Z_14TAGS = t.Z_PK
WHERE
	n.ZARCHIVED=0
	AND n.ZTRASHED=0
	AND lower(t.ZTITLE) LIKE lower('%%%s%%')
ORDER BY
	t.ZMODIFICATIONDATE DESC
LIMIT 25
`
)

type LiteDB struct {
	db *sql.DB
}

func NewLiteDB(path string) (LiteDB, error) {
	db, err := sql.Open("sqlite3", path)
	lite := LiteDB{db}
	return lite, err
}

func (lite LiteDB) Query(q string) ([]map[string]string, error) {
	results := []map[string]string{}
	rows, err := lite.db.Query(q)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return results, err
	}

	for rows.Next() {
		m := map[string]string{}
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return results, err
		}
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			uints, ok := (*val).([]uint8)
			if ok {
				m[colName] = string(uints)
			} else {
				return results, fmt.Errorf("Problem converting record to values to strings for %#v", *val)
			}
		}
		results = append(results, m)
	}
	return results, err
}

func getUniqueTagString(tagString string) string {
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
	return "#" + strings.Join(uniqueTags, " #")
}

func addNoteRowsToAlfred(rows []map[string]string) {
	for _, row := range rows {
		alfred.Add(alfred.Item{
			Title:    row[TitleKey],
			Subtitle: getUniqueTagString(row[TagsKey]),
			Arg:      row[NoteIDKey],
			Valid:    alfred.Bool(true),
		})
	}
}

func getUniqueRows(currentRows, newRows []map[string]string) []map[string]string {
	currentRowIDs := map[string]bool{}
	for _, row := range currentRows {
		currentRowIDs[row[NoteIDKey]] = true
	}

	uniqueRows := []map[string]string{}
	for _, row := range newRows {
		if _, ok := currentRowIDs[row[NoteIDKey]]; !ok {
			uniqueRows = append(uniqueRows, row)
		}
	}

	return uniqueRows
}

func main() {
	query := os.Args[1]

	path := comp.Expanduser(DbPath)
	db, err := NewLiteDB(path)
	if err != nil {
		panic(err)
	}

	elements := strings.Split(query, " ")
	tags := []string{}
	words := []string{}
	lastElement := ""
	for _, e := range elements {
		switch {
		case e == "":
		case strings.HasPrefix(e, "#"):
			tags = append(tags, e)
		default:
			words = append(words, e)
		}
		lastElement = e
	}

	wordStr := strings.Join(words, " ")

	switch {
	case strings.HasPrefix(lastElement, "#"):
		rows, err := db.Query(fmt.Sprintf(TAGS_BY_TITLE, lastElement[1:]))
		if err != nil {
			panic(err)
		}

		for _, row := range rows {
			tag := "#" + row[TitleKey]
			autocomplete := strings.Join(elements[:len(elements)-1], " ") + " " + tag + " "
			alfred.Add(alfred.Item{
				Title:        tag,
				Autocomplete: autocomplete,
				Valid:        alfred.Bool(false),
			})
		}

	case len(words) == 0 && len(tags) == 0 && lastElement == "":
		rows, err := db.Query(RECENT_NOTES)
		if err != nil {
			panic(err)
		}
		addNoteRowsToAlfred(rows)

	case len(tags) != 0:
		tagConditions := []string{}
		for _, t := range tags {
			c := fmt.Sprintf("lower(tag.ZTITLE) = lower('%s')", t[1:])
			tagConditions = append(tagConditions, c)
		}
		tagConjunction := strings.Join(tagConditions, " OR ")
		rows, err := db.Query(fmt.Sprintf(NOTES_BY_TAGS_AND_QUERY, tagConjunction, wordStr, wordStr, len(tags), wordStr))
		if err != nil {
			panic(err)
		}
		addNoteRowsToAlfred(rows)

	default:

		rows, err := db.Query(fmt.Sprintf(NOTES_BY_QUERY, wordStr, wordStr, wordStr))
		if err != nil {
			panic(err)
		}
		addNoteRowsToAlfred(rows)
	}

	alfred.Run()
}
