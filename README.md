# gg

A CLI for your [Gists](gist.github.com). It syncs your gists locally, making them searchable and quickly accessible.

# Getting Started

1. [Create a new authentication token](https://github.com/settings/tokens). Under permissions select 'gist'
2. Run `gg sync --token <authentication_token>`.

# Query Gists

`gg ls` will list your entire gist library in a table. When doing so you will see that every gist is assigned an index. This index can be used to perform operations such as outputting, opening, editing, and removing gists.

# Retrieve Gists

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



# Creating new gists

#### Files

The following will create a new gist that includes both `analysis.R` and `setup.sh`

```bash
gg new --description "analysis scripts" analysis.R setup.sh 
```

#### stdin

You can also pipe input into `gg` to create a new gist, and set gists to be private with `--private`.

```bash
cat analysis_results.tsv | gg new --description "experiment results" --private
```

#### Clipboard

You can create a new gist from your clipboard using:

```bash
gg new --clipboard --description "A new gist" --filename "analysis.sh"
```

# Remove Gists

Use `gg rm` to delete gists. For example:

`gg rm 12`

You can also list multiple gists to remove:

`gg rm 12 134 47`

# Contributing

Feel free to open a PR or suggest changes! Relatively new to Go, so technique suggestions are especially welcome.
