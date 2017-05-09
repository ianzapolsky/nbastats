#!/bin/bash

season_id=$1

cat "data/season_ids/${season_id}_ids.csv" | cut -d, -f2 | xargs -I{} python download_game.py -f csv -o data/${season_id}/{}.csv -i {}
