# Init the all-the-things needed
init:
    pre-commit install
    go get

# Start, migrate, then run with air
dev: start migrate
    go tool air --build.cmd 'go build -tags dev -o ./tmp/main .'

# Search for a decision
decision term:
    rg --ignore-case --multiline --multiline-dotall '> In the context of.*{{ term }}.*\n\n'

# Lint with pre-commit
lint:
    pre-commit run --all-files

# Start with docker compose
start:
    #!/usr/bin/env bash
    docker compose up -d
    echo -n "Waiting for postgres to start..."
    until PGPASSWORD=postgres_pwd psql -h localhost -d postgres -U postgres -c "select 1" &> /dev/null;
    do
        echo -n "."
        sleep 0.1;
    done

# Run sql migrations
migrate:
    PGPASSWORD=skabelon_pwd psql -h localhost -d postgres -U skabelon -f db/api.sql

# Upgrade all tools
upgrade:
    #!/usr/bin/env bash
    pre-commit autoupdate

    # go
    go get -u -t ./...
    go mod tidy
