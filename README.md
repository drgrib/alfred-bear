[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate?business=N2GLXLS5KBFBY&item_name=Chris+Redford&currency_code=USD)

# Bear Workflow

Streamlined note searching and creation for [Bear](http://www.bear-writer.com/) using [Alfred](https://www.alfredapp.com/workflows/).

## Install

Just [download](https://github.com/drgrib/alfred-bear/releases/download/1.2/Bear.alfredworkflow) the latest release and double-click _Bear.alfredworkflow_. Alfred will open the workflow and install it. To avoid getting blocked from running the workflow by the new macOS security features, you can [authorize all the executables used by the workflow](#authorize-all-executables). If you are using an old version of Bear before 2.0 use [this](https://github.com/drgrib/alfred-bear/releases/download/1.1.9/Bear.alfredworkflow) download link for the latest version supporting it.

## Search

`bs` or `bsearch`

### Recent Notes

Leave the search field empty to see recent notes with their tags as subtitles.

<img src="doc/RecentNotes.png" width="500">

### Basic Search

Start typing to search through the titles and text of most recent notes, title matches first.

<img src="doc/BasicSearch.png" width="500">

### Tag Search

Type `#` at any time to autocomplete your tags.

<img src="doc/TagAutocomplete.png" width="500">

Start typing to search tags.

<img src="doc/TagAutocompleteSearch.png" width="500">

Once completed, the notes will be filtered by that tag.

<img src="doc/TagFilter.png" width="500">

Add more tags to filter by multiple tags.

<img src="doc/TagAutocompleteMultiple.png" width="500">

Start typing to search titles and text within a tag.

<img src="doc/TagTextSearch.png" width="500">

All these terms can be typed in any order and they will work the same. For example, if you want to add a tag after typing a bare search term, the autocomplete will still help you. Or if you remember you want to filter by another tag after typing the first tag and a bare search term, you can autocomplete and add the second tag by typing `#` again.

<img src="doc/TagAnyOrder.png" width="500">

### Search in Bear App

You can search any _query_ you type in the Bear app's main window by holding down the option key. If you've entered a tag, it will open the Bear main window to that tag for further browsing.

<img src="doc/SearchQueryInBearApp.png" width="500">

The workflow will also autocomplete any of Bear's [Special Search keywords](https://bear.app/faq/Advanced%20search%20options%20in%20Bear/) if you start typing `@` or `-@`.

<img src="doc/SpecialSearchAutocomplete.png" width="500">

If you use these keywords and have no other search results, the workflow will automatically populate a "Search ... in Bear App" item without you needing to press option.

<img src="doc/SearchSpecialInBearApp.png" width="500">

### Open Note in Bear App

Similarly, you can open any _note_ you select in the Bear app's main window by holding down the command key.

### Link Pasting

While in your Bear notes, you can paste a link to another note by searching for it and holding down the shift key.

<img src="doc/Link1.png" width="500">
<img src="doc/Link2.png" width="500">

### Table of Contents Pasting

Similarly, you can paste a table of contents for a note by searching for it and holding down the command and shift keys.

## New Notes

`bn` or `bnew` followed by title and optional tags.

<img src="doc/NewNote.png" width="500">

Tag autocomplete works the same. Also, any text in your clipboard can be added to the new note by holding down the command key.

## Create/Search

`bcs` or `bcsearch`

You may find sometimes that you want to retrieve a note if it exists and create it if it does not. This command provides that functionality by combining the behavior of search and create. It will provide all the same search results as normal search and additionally add a create item third in the list using normal create options.

<img src="doc/CreateSearch1.png" width="500">

If there are less than two search items, the create item will be the last or only item.

<img src="doc/CreateSearch2.png" width="500">

You can additionally create links to notes by holding the shift key while selecting a search item. Selecting the create item while holding the shift key will do nothing.

## Why I created this

I am especially grateful to Chris Brown, who created a [Python based Bear workflow](https://github.com/chrisbro/alfred-bear). It was the basis for this project. However, I decided to create my own project for a few reasons:

-   Compiled Go is faster than interpretted Python. Not that much faster but fast enough for me to notice when searching and creating notes throughout the day.
-   I wanted the features involving tag searching and autocompletion, link pasting, and automatic clipboard note content.
-   I wanted fewer, more optimized SQL queries into the Bear database to increase speed since this appears to be the main bottleneck on performance.

## Authorization

The first time you use the workflow after installing or upgrading, you will see a security warning:

<img src="doc/Authorize.png" width="400">

This is a quirk of macOS 10.15 and above. Apple currently forces developers to pay $99 a year to be able to officially sign their executables and avoid this warning, which I'm not going to pay since I'm providing this workflow for free as an open source project.

After seeing this warning, you have to go to **System Preferences > Security & Privacy > General** and click the new button that has appeared to allow the executable to run. You then have to run it again and you will see this security warning _again_ but now it will have a new button that lets you allow the executable to run.

These warnings will appear once for each of the executables inside the workflow as you use new features. Once you have authorized all of them, you won't see these warnings anymore until you install a new version.

### Authorize All Executables

If you want to authorize all Bear Workflow executables at once or if you do not see the above security warnings but the workflow isn't working, you can do the following:

1. Go to the **Workflows** section in Alfred Preferences
2. Right click on **Bear** _by drgrib_ and select **Open in Terminal**
3. Copy this command and execute it:

```
xattr -rd com.apple.quarantine cmd
```

This should authorize all the Alfred Bear the executables and fix the security errors.

### 🍎 Apple Silicon Macs

If your mac is based on Apple Silicon chip, you need to have Rosetta installed on your system, otherwise Alfred workflows will fail silently.

<img width="582" alt="Screen Shot 2021-12-29 at 9 06 02 AM" src="https://user-images.githubusercontent.com/9834975/147670554-eae2ca66-b929-4a03-b59e-545d3e660082.png">

Copy this command and execute in terminal to install Rosetta:

```sh
softwareupdate --install-rosetta
```
