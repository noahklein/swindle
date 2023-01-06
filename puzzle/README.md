# Puzzle

Test the engine against the [Lichess Puzzle DB](https://database.lichess.org/#puzzles).

```bash
# Download the DB.
./download.sh
```

## Examples
```bash
go run bin/main.go -tag mateIn3 -limit 20
go run bin/main.go -id 00a98

# Get a list of all tags.
go run bin/main.go -tsearch '*'

# Get a list of mate tags.
go run bin/main.go -tsearch 'mate'
```
