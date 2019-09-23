# Bear Workflow
Streamlined note searching and creation for [Bear](http://www.bear-writer.com/) using [Alfred](https://www.alfredapp.com/workflows/).

## Install
Just download the latest release and double-click _Bear.alfredworkflow_. Alfred will open the workflow and install it.

## Search
`bs` or `bsearch`

### Recent Notes
Leave the search field empty to see recent notes with their tags as subtitles.

![](doc/RecentNotes.png | width=100)

### Basic Search
Start typing to search through the titles and text of most recent notes, title matches first.

![](doc/BasicSearch.png | width=100)

### Tag Search
Type `#` at any time to autocomplete your tags.

![](doc/TagAutocomplete.png | width=100)

Start typing to search tags.

![](doc/TagAutocompleteSearch.png | width=100)

Once completed, the notes will be filtered by that tag.

![](doc/TagFilter.png | width=100)

Add more tags to filter by multiple tags.

![](doc/TagAutocompleteMultiple.png | width=100)

Start typing to search titles and text within a tag.

![](doc/TagTextSearch.png | width=100)

All these terms can be typed in any order and they will work the same. For example, if you want to add a tag after typing a bare search term, the autocomplete will still help you. Or if you remember you want to filter by another tag after typing the first tag and a bare search term, you can autocomplete and add the second tag by typing `#` again.

![](doc/TagAnyOrder.png | width=100)

### Link Pasting

While in your Bear notes, you can paste a link to another note by searching for it and holding down the command key.

![](doc/Link1.png | width=100)
![](doc/Link2.png | width=100)

## New Notes
`bn` or `bnew` followed by title and optional tags.

![](doc/NewNote.png | width=100)

Tag autocomplete works the same. Also, any text in your clipboard will automatically be added to the new note. Finally, the new note will be opened on creation unless you hold down the command key.

## Why this was created
I am especially grateful to Chris Brown, who created a [Python based Bear workflow](https://github.com/chrisbro/alfred-bear). It was the basis for this project. However, I decided to create my own project for a few reasons:

- Compiled Go is faster than interpretted Python. Not that much faster but fast enough for me to notice when searching and creating notes throughout the day.
- I wanted the features involving tag searching and autocompletion, link pasting, and automatic clipboard note content.
- I wanted fewer, more optimized SQL queries into the Bear database to increase speed since this appears to be the main bottleneck on performance.
