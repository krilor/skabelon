# Run with air
dev:
    go tool air

decision term:
    rg --ignore-case --multiline --multiline-dotall '> In the context of.*{{ term }}.*\n\n'
