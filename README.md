# gg

A CLI for your [Gists](gist.github.com). It syncs your gists locally, making them searchable and quickly accessible.

# Features

* Syntax highlighting in terminal
* Create, edit, and delete gists from the command line
* Edit gists in a text editor (e.g. sublime)
* Organize gists using tags (`#hashtag` syntax)
* Full text search
* Filter and sort by tag, language, owner, public/private, starred, search term
* `gg` includes starred gists by other users
* Summarize gists by tag, language, or owner

<!--# Demo-->
# Usage

## Getting Started

1. [Create a new authentication token](https://github.com/settings/tokens). Under permissions select 'gist'
2. Run `gg sync --token <authentication_token>`.

## Query Gists

`gg ls` can be used to search and filter your gist library. Results are output in a table. For convenience, the `ls` command is run implicitly when `gg` is invoked as long as the term being passed is not also a command.

For example:

```bash
gg # Lists recent gists
gg ls # equivelent to above

gg genomics # Searches gists for the term 'genomics'
gg ls genomics # equivelent to above

gg -l 100 # lists last 100 gists.
gg ls -l 100 # lists the last 100 gists

gg help # shows help
gg sync # syncs gists
```

The `gg` command list is:

* `#` - any integer number.
* `help`, `h`, `--help`, `-h`
* `sync`
* `set-editor` 
* `logout`
* `new`
* `edit`
* `web`, `w`
* `open`, `o`
* `rm`
* `ls`, `list`
* `search`
* `starred`
* `tag`, `tags`
* `language`, `languages`
* `owner`
* `__run_alfred`, `debug`

You can still search on these terms by using `ls` explicitly:

```shell
gg sync # runs the sync command
gg ls sync # searches for the term 'sync'
```

![Gist List](https://github.com/danielecook/gg/blob/media/gist_list.png?raw=true)

## Retrieve Gists

```bash
gg open 5 # Outputs a single gist
gg o 5 # 'o' is a shortcut for open.

# To be even quicker, gg will open a gist when the first argument is an integer.
gg 5 # equivelent to `gg o 5` or `gg open 5`

# Output multiple gists
gg 5 8 22

# You can pipe the contents to be evaluated; They will not be syntax-highlighted
gg 5 | sh
```

![Gist Retrieval](https://github.com/danielecook/gg/blob/media/syntax.png?raw=true)

## Summarize gists

Gists can be summarized by tag, owner, and language.

```bash
gg tags # Summarize by tag
gg tags fastq # Query gists tagged with 'fastq'
gg tags fastq elegans # Query gists tagged with 'fastq' and containing the word 'elegans'
gg owner # List owners and count
gg language # Table languages and count
```

![summary output](https://github.com/danielecook/gg/blob/media/summary.png?raw=true)

## Creating new gists

#### Files

The following will create a new gist that includes both `analysis.R` and `setup.sh`

```bash
gg new --description "analysis scripts" analysis.R setup.sh 
```

#### stdin

You can also pipe input into `gg` to create a new gist, and set gists to `--private`

```bash
cat analysis_results.tsv | gg new --description "experiment results" --private
```

#### Clipboard

You can create a new gist from your clipboard

```bash
gg new --clipboard --description "A new gist" --filename "analysis.sh"
```

## Edit Gists

Use `gg edit` to edit gists.

```bash
gg edit 12 # Open gist in text file for editing.
```

Editing a gist opens a text document with the following header:

```
# GIST FORM: Edit Metadata below
# ==============================
# description: tmux shortcuts & cheatsheet
# starred: T
# public: T
# ==============================
```

You can modify the description, starred (T=true; F=false), and public (T=true; F=false) in this header.

Following the header you will see a special separator line that looks like this:

```
myscript.sh----------------------------------------------------------------------------::>>>
```

Because gists can have multiple files, they are represented in a single file by breaking them up using a  specialized line. The line **must** begin with the filename, followed by dashes `---`, and finally end with `::>>>`. A gist with two files looks like this:

```markdown
# GIST FORM: Edit Metadata below
# ==============================
# description: analysis.R
# starred: T
# public: T
# ==============================
README.md-----------------------------------------------------------------------------::>>>
I used R to analyze my data!

Check out analysis.R to see how I did it.
analysis.R----------------------------------------------------------------------------::>>>
# My R script
print(1 + 1)
```

Supported editors:

* Sublime Text (`subl`)
* Nano

## Remove Gists

Use `gg rm` to delete gists.

```bash
gg rm 12
gg rm 12 134 47 # Remove multiple gists
```

# Contributing

Feel free to open a PR or suggest changes! Relatively new to Go, so technique suggestions are especially welcome.
