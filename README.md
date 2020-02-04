# gg

A command-line utility for accessing gists. Also an Alfred Workflow

# Getting Started

Run `gg login <authtoken>`.

# Creating new gists

## Files

The following will create a new gist that includes both `analysis.R` and `setup.sh`

```bash
gg new --description "analysis scripts" analysis.R setup.sh 
```

## stdin

You can also pipe input into `gg` to create a new gist. You can also set to private with `--private`.

```bash
cat analysis_results.tsv | gg new --description "experiment results" --private
```

## Clipboard

You can create a new gist from your clipboard using:

```bash
gg new --clipboard --description "A new gist" --filename "analysis.sh"
```

# Listing Gists

`gg ls` will list your entire gist library in a table.

# Retrieve a specific gist

# Remove Gists

Use `gg rm` to delete gists. For example:

`gg rm 12`

You can also list multiple gists to remove:

`gg rm 12 134 47`


