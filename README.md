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

`gg ls` will list your gist library in a table. Each gist in your library is assigned an ID which can be used when performing operations such as opening and editing gists.

```bash
gg ls --language Shell # Filter gists by language Shell
gg ls --tag database # Filter gists with the tag 'database'
gg ls FASTQ # Search gists with the word 'FASTQ'
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

## Remove Gists

Use `gg rm` to delete gists.

```bash
gg rm 12
gg rm 12 134 47 # Remove multiple gists
```

# Contributing

Feel free to open a PR or suggest changes! Relatively new to Go, so technique suggestions are especially welcome.
