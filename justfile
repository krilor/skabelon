init:
    pre-commit install
    go get

# Run with air
dev: start migrate
    go tool air --build.cmd 'go build -tags dev -o ./tmp/main .'

decision term:
    rg --ignore-case --multiline --multiline-dotall '> In the context of.*{{ term }}.*\n\n'

lint:
    pre-commit run --all-files

start:
    #!/usr/bin/env bash
    docker compose up -d
    echo -n "Waiting for postgres to start..."
    until PGPASSWORD=postgres_pwd psql -h localhost -d postgres -U postgres -c "select 1" &> /dev/null;
    do
        echo -n "."
        sleep 0.1;
    done

migrate:
    PGPASSWORD=skabelon_pwd psql -h localhost -d postgres -U skabelon -f db/henhold.sql

upgrade:
    #!/usr/bin/env bash
    pre-commit autoupdate

    # go
    go get -u -t ./...
    go mod tidy
