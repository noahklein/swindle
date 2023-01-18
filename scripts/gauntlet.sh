#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
EDIR="$SCRIPT_DIR/competitors"

cutechess-cli -tournament gauntlet -concurrency 2 \
    -engine cmd="$EDIR/swindle" \
    -engine cmd="$EDIR/rustic3" \
    -engine cmd="$EDIR/stockfish15" \
    -engine cmd="$EDIR/counter5" \
    -each tc=40/20 proto=uci option.Hash=32 \
    -resign movecount=3 score=500 \
    -draw movenumber=50 movecount=5 score=20 \
    -games 2 -rounds "$1" \
