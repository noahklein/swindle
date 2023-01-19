#!/usr/bin/env bash

# Play stockfish using the UCI_LimitStrength option.

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
EDIR="$SCRIPT_DIR/competitors"

TIME_CONTROL=$1
ELO=$2

cutechess-cli \
    -rounds 4 -concurrency 2 \
    -engine cmd="$EDIR/swindle" \
    -engine cmd="$EDIR/stockfish15" option.UCI_LimitStrength=true option.UCI_Elo="$ELO" \
    -each tc="$TIME_CONTROL" proto=uci option.Hash=32 \
    -sprt elo0=0 elo1=10 alpha=0.05 beta=0.05 \
    -pgnout 'tournament.pgn'
