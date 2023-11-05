package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/drgrib/alfred-bear/db"
)

type TagQueryArg struct {
	Text           string
	IntersectQuery string
}

func TemplateToString(templateStr string, data any) (string, error) {
	var buffer bytes.Buffer
	t := template.Must(template.New("").Parse(templateStr))
	err := t.Execute(&buffer, data)
	return buffer.String(), err
}

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	litedb, err := db.NewBearDB()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	targetTags := []string{"#health/sleep", "#media/youtube"}
	notes, err := litedb.QueryNotesByTextAndTags("", targetTags)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	for _, n := range notes {
		tags := strings.Split(n[db.TagsKey], ",")
		found := map[string]bool{}
		for _, t := range tags {
			found[t] = true
		}
		invalid := false
		for _, t := range targetTags {
			if !found[db.RemoveTagHashes(t)] {
				invalid = true
			}
		}
		if invalid {
			fmt.Print("INVALID ")
		}
		fmt.Println(n[db.TitleKey], n[db.TagsKey])
	}

	fmt.Println()

	queryTemplate := `
	WITH 
		joined AS (
			SELECT
				note.ZUNIQUEIDENTIFIER,
				note.ZTITLE,
				note.ZTEXT,
				note.ZMODIFICATIONDATE,
				tag.ZTITLE AS TAG_TITLE,
				images.ZSEARCHTEXT
			FROM
				ZSFNOTE note
			INNER JOIN
				Z_5TAGS nTag ON note.Z_PK = nTag.Z_5NOTES
			INNER JOIN
				ZSFNOTETAG tag ON nTag.Z_13TAGS = tag.Z_PK
			LEFT JOIN
				ZSFNOTEFILE images ON images.ZNOTE = note.Z_PK
			WHERE
				note.ZARCHIVED = 0
				AND note.ZTRASHED = 0
				AND note.ZTEXT IS NOT NULL
		),
		hasSearchedTags AS (
			{{ .IntersectQuery}}
		)
	SELECT
		ZUNIQUEIDENTIFIER,
		ZTITLE,
		GROUP_CONCAT(DISTINCT TAG_TITLE) AS TAGS
	FROM
		joined
	WHERE
		ZUNIQUEIDENTIFIER IN hasSearchedTags 
		AND (
			utflower(ZTITLE) LIKE utflower('%{{ .Text}}%') OR
			utflower(ZTEXT) LIKE utflower('%{{ .Text}}%') OR
			ZSEARCHTEXT LIKE utflower('%{{ .Text}}%')
		)
	GROUP BY
		ZUNIQUEIDENTIFIER,
		ZTITLE
	ORDER BY
		CASE WHEN utflower(ZTITLE) LIKE utflower('%{{ .Text}}%') THEN 0 ELSE 1 END,
		ZMODIFICATIONDATE DESC
	LIMIT 20
	`

	var selectStatements []string
	for _, t := range targetTags {
		s := fmt.Sprintf("SELECT ZUNIQUEIDENTIFIER FROM joined WHERE utflower(TAG_TITLE) = utflower('%s')", db.RemoveTagHashes(t))
		selectStatements = append(selectStatements, s)
	}

	tagQueryArg := TagQueryArg{
		Text:           "",
		IntersectQuery: strings.Join(selectStatements, "\nINTERSECT\n"),
	}

	query, err := TemplateToString(queryTemplate, tagQueryArg)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	notes, err = litedb.Query(query)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	for _, n := range notes {
		fmt.Println(n[db.TitleKey], n[db.TagsKey])
	}

	fmt.Println(tagQueryArg.IntersectQuery)
}
