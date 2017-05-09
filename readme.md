## get all players
cat * | awk -F ',' '{ print $14 }' | sort | uniq > players.dat
