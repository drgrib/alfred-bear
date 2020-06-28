package db

import (
	"database/sql"
	"fmt"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
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
	LEFT OUTER JOIN Z_7TAGS nTag ON note.Z_PK = nTag.Z_7NOTES
	LEFT OUTER JOIN ZSFNOTETAG tag ON nTag.Z_14TAGS = tag.Z_PK
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
	LEFT OUTER JOIN Z_7TAGS nTag ON note.Z_PK = nTag.Z_7NOTES
	LEFT OUTER JOIN ZSFNOTETAG tag ON nTag.Z_14TAGS = tag.Z_PK
WHERE
	note.ZARCHIVED=0
	AND note.ZTRASHED=0
	AND (
		lower(note.ZTITLE) LIKE lower('%%%s%%') OR
		lower(note.ZTEXT) LIKE lower('%%%s%%')
	)
GROUP BY note.ZUNIQUEIDENTIFIER
ORDER BY case when lower(note.ZTITLE) LIKE lower('%%%s%%') then 0 else 1 end, note.ZMODIFICATIONDATE DESC
LIMIT 200
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
LIMIT 200
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

	NOTE_TITLE_BY_ID = `
SELECT DISTINCT
    ZTITLE
FROM
    ZSFNOTE
WHERE
    ZARCHIVED=0
    AND ZTRASHED=0
    AND ZUNIQUEIDENTIFIER='%s'
ORDER BY
    ZMODIFICATIONDATE DESC
LIMIT 25
`
)

type Note map[string]string

func Expanduser(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path[:2] == "~/" {
		path = filepath.Join(dir, path[2:])
	}
	return path
}

type LiteDB struct {
	db *sql.DB
}

func NewLiteDB(path string) (LiteDB, error) {
	db, err := sql.Open("sqlite3", path)
	litedb := LiteDB{db}
	return litedb, err
}

func NewBearDB() (LiteDB, error) {
	path := Expanduser(DbPath)
	litedb, err := NewLiteDB(path)
	return litedb, err
}

func (litedb LiteDB) Query(q string) ([]Note, error) {
	results := []Note{}
	rows, err := litedb.db.Query(q)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return results, err
	}

	for rows.Next() {
		m := Note{}
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
			s, ok := (*val).(string)
			if ok {
				m[colName] = s
			} else {
				m[colName] = ""
			}
		}
		results = append(results, m)
	}
	return results, err
}

func escape(s string) string {
	return strings.Replace(s, "'", "''", -1)
}

func containsOrderedWords(text string, words []string) bool {
	prev := 0
	for _, w := range words {
		i := strings.Index(text, w)
		if i == -1 || i < prev {
			return false
		}
		prev = i
	}
	return true
}

func containsWords(text string, words []string) bool {
	for _, w := range words {
		if !strings.Contains(text, w) {
			return false
		}
	}
	return true
}

func (litedb LiteDB) queryNotesByTextAndTagConjunction(text, tagConjunction string, tags []string) ([]Note, error) {
	text = escape(text)
	return litedb.Query(fmt.Sprintf(NOTES_BY_TAGS_AND_QUERY, tagConjunction, text, text, len(tags), text))
}

func (litedb LiteDB) QueryNotesByTextAndTags(text string, tags []string) ([]Note, error) {
	tagConditions := []string{}
	for _, t := range tags {
		c := fmt.Sprintf("lower(tag.ZTITLE) = lower('%s')", t[1:])
		tagConditions = append(tagConditions, c)
	}
	tagConjunction := strings.Join(tagConditions, " OR ")

	wordQuery := func(word string) ([]Note, error) {
		return litedb.queryNotesByTextAndTagConjunction(word, tagConjunction, tags)
	}

	return multiWordQuery(text, wordQuery)
}

func (litedb LiteDB) QueryNotesByText(text string) ([]Note, error) {
	wordQuery := func(word string) ([]Note, error) {
		word = escape(word)
		return litedb.Query(fmt.Sprintf(NOTES_BY_QUERY, word, word, word))
	}
	return multiWordQuery(text, wordQuery)
}

func splitSpacesOrQuoted(s string) []string {
	r := regexp.MustCompile(`([^\s"']+)|"([^"]*)"`)
	matches := r.FindAllStringSubmatch(s, -1)
	var split []string
	for _, v := range matches {
		match := v[1]
		if match == "" {
			match = v[2]
		}
		split = append(split, match)
	}
	return split
}

type noteRecord struct {
	note                 Note
	contains             bool
	containsOrderedWords bool
	containsWords        bool
	originalIndex        int
}

func NewNoteRecord(i int, note Note, lowerText string) *noteRecord {
	title := strings.ToLower(note[TitleKey])
	words := strings.Split(lowerText, " ")
	record := noteRecord{
		originalIndex:        i,
		note:                 note,
		contains:             strings.Contains(title, lowerText),
		containsOrderedWords: containsOrderedWords(title, words),
		containsWords:        containsWords(title, words),
	}
	return &record
}

func multiWordQuery(text string, wordQuery func(string) ([]Note, error)) ([]Note, error) {
	lowerText := strings.ToLower(text)
	words := splitSpacesOrQuoted(lowerText)

	var noteRecords []*noteRecord
	count := map[string]int{}
	for _, word := range words {
		notes, err := wordQuery(word)
		if err != nil {
			return nil, err
		}

		for i, note := range notes {
			noteId := note[NoteIDKey]
			if count[noteId] == 0 {
				record := NewNoteRecord(i, note, lowerText)
				record.originalIndex = i
				noteRecords = append(noteRecords, record)
			}
			count[noteId]++
		}
	}

	var finalRecords []*noteRecord
	for _, record := range noteRecords {
		if count[record.note[NoteIDKey]] == len(words) || record.containsWords {
			finalRecords = append(finalRecords, record)
		}
	}

	sort.Slice(finalRecords, func(i, j int) bool {
		iRecord := finalRecords[i]
		jRecord := finalRecords[j]

		if iRecord.contains != jRecord.contains {
			return iRecord.contains
		}

		if iRecord.containsOrderedWords != jRecord.containsOrderedWords {
			return iRecord.containsOrderedWords
		}

		if iRecord.containsWords != jRecord.containsWords {
			return iRecord.containsWords
		}

		return iRecord.originalIndex < jRecord.originalIndex
	})

	var finalRows []Note
	for _, noteRecord := range finalRecords {
		finalRows = append(finalRows, noteRecord.note)
	}

	return finalRows, nil
}
