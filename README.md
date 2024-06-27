# GoComplain

[![Yum](https://img.shields.io/badge/-Buy%20me%20a%20cookie-blue?labelColor=grey&logo=cookiecutter&style=for-the-badge)](https://www.buymeacoffee.com/mjwhitta)

[![Go Report Card](https://goreportcard.com/badge/github.com/mjwhitta/gocomplain?style=for-the-badge)](https://goreportcard.com/report/github.com/mjwhitta/gocomplain)
![License](https://img.shields.io/github/license/mjwhitta/gocomplain?style=for-the-badge)

## What is this?

This module attempts to combine multiple other Go source analyzing
tools. Currently supported functionality includes:

- Go analyzers
    - go vet
    - gocyclo
    - golint
    - ineffassign
    - staticcheck
- Go formatting
    - gofmt
    - gofumpt
    - line-length verification
- Spelling
    - misspell
    - spellcheck

## How to install

Open a terminal and run the following:

```
$ go install github.com/mjwhitta/gocomplain/cmd/gocomplain@latest
```

## Usage

Run `gocomplain -h` to see the full usage, but you can safely run
`gocomplain` to get started analyzing, while using the default
settings.
