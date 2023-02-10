# Swindle
The Swindle chess engine.

## Installation
Download the [latest release](https://github.com/noahklein/swindle/releases/latest).

Or [install go](https://go.dev/doc/install) and run/build from source:
```bash
# Run
go run .

# Build and run
go build -o swindle
./swindle
```

## Usage
Swindle is [UCI-compatible](https://www.wbec-ridderkerk.nl/html/UCIProtocol.html) and can be used as a CLI or in a Chess GUI.
I recommend using [Cute Chess](https://github.com/cutechess/cutechess), but here's a [list of popular UCI-compatible GUIs](https://www.chessprogramming.org/UCI#GUIs).


### CLI Reference
```
help            Print the help dialogue with example commands.
uci             Print information about Swindle along with available options for the setoption command.
setoption       Set engine options. Print available options with the uci command.
ucinewgame      Starts a new game, resets engine.
position        Set the current position.
go              Search the position and report the best move.
stop            Cancel search and report the best move the engine has found so far.
exit            Exit the program
```

### Example
For a full list of comamnds, read the UCI protocol: http://wbec-ridderkerk.nl/html/UCIProtocol.html.
```
# Set-up a position:
ucinewgame
position startpos moves e2e4 e7e5
  
# To search a position at depth 10:
go depth 10

# Infinite search:
go infinite
stop
```

## Major Features
### Search
* [Negamax + AlphaBeta Pruning](https://en.wikipedia.org/wiki/Negamax#Negamax_with_alpha_beta_pruning)
* [Quiescence Search](https://www.chessprogramming.org/Quiescence_Search)
* [Transposition Table](https://www.chessprogramming.org/Transposition_Table)
* [Iterative Deepening](https://www.chessprogramming.org/Iterative_Deepening)
* [Aspiration Window](https://www.chessprogramming.org/Aspiration_Windows)
* [Principal Variation Search](https://www.chessprogramming.org/Principal_Variation_Search)
* [Null-move pruning](https://www.chessprogramming.org/Null_Move_Pruning)
* [Delta Pruning](https://www.chessprogramming.org/Delta_Pruning)
* [Futility Pruning](https://www.chessprogramming.org/Futility_Pruning)
* [Late Move Reductions](https://www.chessprogramming.org/Late_Move_Reductions)
* [MVV-LVA Move Ordering](https://www.chessprogramming.org/MVV-LVA)

### Evaluation
* [Tapered Eval](https://www.chessprogramming.org/Tapered_Eval)
* [Piece-Square Tables](https://www.chessprogramming.org/Piece-Square_Tables)


## [Puzzle package](https://github.com/noahklein/swindle/tree/master/puzzle)
Let's you download and query the [Lichess Puzzle DB](https://database.lichess.org/#puzzles) for engine testing.
