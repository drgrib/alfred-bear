package db

import (
	"database/sql"
	"fmt"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

const (
	DbPath = "~/Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data/database.sqlite"

	TitleKey  = "ZTITLE"
	TextKey   = "ZTEXT"
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
	AND note.ZTEXT IS NOT NULL
	AND (
		utflower(note.ZTITLE) LIKE utflower('%'||$1||'%') OR
		utflower(note.ZTEXT) LIKE utflower('%'||$1||'%')
	)
GROUP BY note.ZUNIQUEIDENTIFIER
ORDER BY case when utflower(note.ZTITLE) LIKE utflower('%'||$1||'%') then 0 else 1 end, note.ZMODIFICATIONDATE DESC
LIMIT 400
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
		AND note.ZTEXT IS NOT NULL
		AND (%s)
		AND (
			utflower(note.ZTITLE) LIKE utflower('%%%s%%') OR
			utflower(note.ZTEXT) LIKE utflower('%%%s%%')
		)
	GROUP BY note.ZUNIQUEIDENTIFIER
	HAVING COUNT(*) >= %d
)
GROUP BY note.ZUNIQUEIDENTIFIER
ORDER BY case when utflower(note.ZTITLE) LIKE utflower('%%%s%%') then 0 else 1 end, note.ZMODIFICATIONDATE DESC
LIMIT 400
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
	AND utflower(t.ZTITLE) LIKE utflower('%%%s%%')
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
	NOTE_TEXT_BY_ID = `
SELECT DISTINCT
    ZTEXT
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
	sql.Register("sqlite3_custom", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			return conn.RegisterFunc("utflower", utfLower, true)
		},
	})

	db, err := sql.Open("sqlite3_custom", path)
	litedb := LiteDB{db}
	return litedb, err
}

func utfLower(s string) string {
	return strings.ToLower(s)
}

func NewBearDB() (LiteDB, error) {
	path := Expanduser(DbPath)
	litedb, err := NewLiteDB(path)
	return litedb, err
}

func (litedb LiteDB) Query(q string, args ...interface{}) ([]Note, error) {
	results := []Note{}
	rows, err := litedb.db.Query(q, args...)
	if err != nil {
		return results, errors.WithStack(err)
	}

	cols, err := rows.Columns()
	if err != nil {
		return results, errors.WithStack(err)
	}

	for rows.Next() {
		m := Note{}
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return results, errors.WithStack(err)
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
	err = rows.Close()
	if err != nil {
		return results, errors.WithStack(err)
	}
	err = rows.Err()
	if err != nil {
		return results, errors.WithStack(err)
	}
	return results, errors.WithStack(err)
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
		c := fmt.Sprintf("utflower(tag.ZTITLE) = utflower('%s')", t[1:])
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
		return litedb.Query(NOTES_BY_QUERY, word)
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

	var records []*noteRecord
	fullMatch := map[string]bool{}
	notes, err := wordQuery(lowerText)
	if err != nil {
		return nil, err
	}
	for i, note := range notes {
		noteId := note[NoteIDKey]
		record := NewNoteRecord(i, note, lowerText)
		record.originalIndex = i
		records = append(records, record)
		fullMatch[noteId] = true
	}

	var multiRecords []*noteRecord
	count := map[string]int{}
	for _, word := range words {
		notes, err := wordQuery(word)
		if err != nil {
			return nil, err
		}

		for i, note := range notes {
			noteId := note[NoteIDKey]
			if count[noteId] == 0 && !fullMatch[noteId] {
				record := NewNoteRecord(i, note, lowerText)
				record.originalIndex = i
				multiRecords = append(multiRecords, record)
			}
			count[noteId]++
		}
	}

	for _, record := range multiRecords {
		if count[record.note[NoteIDKey]] == len(words) || record.containsWords {
			records = append(records, record)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		iRecord := records[i]
		jRecord := records[j]

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
	for _, noteRecord := range records {
		finalRows = append(finalRows, noteRecord.note)
	}

	return finalRows, nil
}
