#!/usr/bin/env bash

PUZZLE_ZIP='lichess_db_puzzle.csv.zst'

curl https://database.lichess.org/lichess_db_puzzle.csv.zst > $PUZZLE_ZIP
zstd -d $PUZZLE_ZIP
rm $PUZZLE_ZIP