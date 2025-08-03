module github.com/undeconstructed/skribserv

go 1.23.0

replace github.com/go-rel/postgres => ../rel-postgres

require (
	github.com/PumpkinSeed/slog-context v0.1.2
	github.com/go-rel/migration v0.3.1
	github.com/go-rel/postgres v0.12.0
	github.com/go-rel/rel v0.42.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/phsym/console-slog v0.3.1
	github.com/stretchr/testify v1.9.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-rel/sql v0.17.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
