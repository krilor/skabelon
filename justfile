init:
    pre-commit install
    go get

# Run with air
dev:
    go tool air --build.cmd 'go build -tags dev -o ./tmp/main .'

decision term:
    rg --ignore-case --multiline --multiline-dotall '> In the context of.*{{ term }}.*\n\n'

lint:
    pre-commit run --all-files

start:
    docker compose up -d

upgrade:
    #!/usr/bin/env bash
    pre-commit autoupdate

    # go
    go get -u -t ./...
    go mod tidy
