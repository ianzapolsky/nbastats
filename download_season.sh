#!/bin/bash

season_id=$1

mkdir -p data/${season_id}_playoffs
cat "data/playoff_ids/${season_id}_ids.csv" | cut -d, -f2 | xargs -I{} python download_game.py -f csv -o data/${season_id}_playoffs/{}.csv -i {}
