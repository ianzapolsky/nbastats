package main

import (
	"log"
	"strings"
)

const HOT_SEC int = 60 * 2

func getHotColdHeaderRow() []string {
	return []string{
		"Name",
		"Total 3P Made",
		"Total 3P Att",
		"Total 3P%",
		"Cold 3P Made",
		"Cold 3P Att",
		"Cold 3P%",
		"Hot 3P Made",
		"Hot 3P Att",
		"Hot 3P%",
		"Hot/Cold Diff",
		"Hot Shot Makeup",
	}
}

func collectHotColdStatsFromSeasonForPlayer(m map[string]*playerData, season *Season, player string) {
	log.Printf("collecting stats for %v from season %v", player, season.Id)

	hotChances := 0
	hotMakes := 0
	coldChances := 0
	coldMakes := 0

	for _, game := range season.Games {

		gameHotChances := 0
		gameHotMakes := 0
		gameColdChances := 0
		gameColdMakes := 0
		last3PM := 0

		for _, event := range game {
			if strings.Contains(event.Player, player) {
				if event.Is3PA() {

					if last3PM == 0 {
						gameColdChances += 1
						if event.Is3PM() {
							gameColdMakes += 1
							last3PM = event.TimeSec
						}
					} else {
						if event.TimeSec <= (last3PM + HOT_SEC) {
							gameHotChances += 1
							if event.Is3PM() {
								gameHotMakes += 1
								last3PM = event.TimeSec
							}
						} else {
							gameColdChances += 1
							if event.Is3PM() {
								gameColdMakes += 1
								last3PM = event.TimeSec
							}
						}
					}

				}
			}
		}
		hotChances += gameHotChances
		hotMakes += gameHotMakes
		coldChances += gameColdChances
		coldMakes += gameColdMakes
	}

	totalChances := hotChances + coldChances
	totalMakes := hotMakes + coldMakes

	playerData := &playerData{
		Name:         player,
		TotalMakes:   float64(totalMakes),
		TotalChances: float64(totalChances),
		ColdMakes:    float64(coldMakes),
		ColdChances:  float64(coldChances),
		HotMakes:     float64(hotMakes),
		HotChances:   float64(hotChances),
	}

	lookup, ok := m[player]
	if !ok {
		m[player] = playerData
	} else {
		lookup.TotalMakes += playerData.TotalMakes
		lookup.TotalChances += playerData.TotalChances
		lookup.ColdMakes += playerData.ColdMakes
		lookup.ColdChances += playerData.ColdChances
		lookup.HotMakes += playerData.HotMakes
		lookup.HotChances += playerData.HotChances
	}
}
