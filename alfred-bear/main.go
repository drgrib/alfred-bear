package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/drgrib/alfred"
)

func main() {
	query := os.Args[1]

	elements := strings.Split(query, " ")
	tags := []string{}
	words := []string{}
	for _, e := range elements {
		switch {
		case e == "":
		case strings.HasPrefix(e, "#"):
			tags = append(tags, e)
		default:
			words = append(words, e)
		}
	}

	alfred.Add(alfred.Item{
		Title:    query,
		Subtitle: fmt.Sprintf("%v %v %v %v", words, len(words), tags, len(tags)),
		Arg:      "arg",
		UID:      "uid",
	})

	alfred.Run()
}
