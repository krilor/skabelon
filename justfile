# Run with air
dev:
    go tool air --build.cmd 'go build -tags dev -o ./tmp/main .'

decision term:
    rg --ignore-case --multiline --multiline-dotall '> In the context of.*{{ term }}.*\n\n'
