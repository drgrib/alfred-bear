package db

import (
	"database/sql"
	"html"
	"os/user"
	"path/filepath"

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

func (lite LiteDB) Query(q string) ([]map[string]string, error) {
	results := []map[string]string{}
	rows, err := lite.db.Query(html.EscapeString(q))
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
				m[colName] = ""
			}
		}
		results = append(results, m)
	}
	return results, err
}
