# gotrain

Boilerplate generator for Go web-apps. Uses GORM, SQLite and HTMX to stand up an MVP in seconds.

## Requirements

- Go 1.25
- git

## Quickstart

`go install github.com/Angus-Warman/gotrain`
`mkdir my-gotrain-app`
`cd my-gotrain-app`
`gotrain create`
`gotrain generate model user username:required:unique email:required:unique`
`gotrain generate migration`
`gotrain run migration`
`go run .`
