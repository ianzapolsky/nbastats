#!/bin/bash

# this script downloads the play-by-play data for every game in the regular season
# and playoffs for a given year, in separate directories.

if [ -z $1 ]; then
  echo "Usage: ./download_season.sh <season_id>"
  echo "\te.g. ./download_season.sh 2016-17"
  exit 1
fi

season_id=$1
playoffs_id=${season_id}_playoffs

season_dir=data/${season_id}
playoffs_dir=data/${playoffs_id}

mkdir -p $season_dir
mkdir -p $playoffs_dir

season_ids_file=/tmp/${season_id}_ids
playoffs_ids_file=/tmp/${playoffs_id}_ids

# This gets the list of game IDs for each season
python get_game_ids.py -s ${season_id} -o ${season_ids_file} -f csv -t 'Regular Season'
python get_game_ids.py -s ${season_id} -o ${playoffs_ids_file} -f csv -t 'Playoffs'

# Now, iterate over list of game IDs for each season and download
cat $season_ids_file | cut -d, -f2 | xargs -I{} python download_game.py -f csv -o $season_dir/{}.csv -i {}
cat $playoffs_ids_file | cut -d, -f2 | xargs -I{} python download_game.py -f csv -o $playoffs_dir/{}.csv -i {}

# Remove temporary game ID lists
rm $season_ids_file $playoffs_ids_file

# Build player database for each season
cd $season_dir
cat * | awk -F ',' '{ print $14 }' | sort | uniq > players.dat
cd $playoffs_dir
cat * | awk -F ',' '{ print $14 }' | sort | uniq > players.dat
