package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"strings"
)

const (
	DATA_DIR string = "data"
)

var outputFile string
var seasonIdsString string
var seasonIds []string

func main() {

	flag.StringVar(&outputFile, "o", "", "output file")
	flag.StringVar(&seasonIdsString, "s", "", "season ids (comma separated)")
	flag.Parse()
	if outputFile == "" {
		log.Fatalf("Must supply -o")
	}
	if seasonIdsString == "" {
		log.Fatalf("Must supply -s")
	}
	seasonIds = strings.Split(seasonIdsString, ",")

	seasons := make([]*Season, len(seasonIds))
	for i, sid := range seasonIds {
		season, err := NewSeason(sid)
		if err != nil {
			log.Fatalf("Error creating new season %v: %v", sid, err)
		}
		log.Printf("built season %s", sid)
		seasons[i] = season
	}

	of, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer of.Close()

	writer := csv.NewWriter(of)
	defer writer.Flush()

	masterMap := make(map[string]*playerData)

	// Run the hot/cold report, the logic for which is defined in hcreport.go
	for _, season := range seasons {
		for _, player := range season.Players {
			collectHotColdStatsFromSeasonForPlayer(masterMap, season, player)
		}
	}

	for _, p := range masterMap {
		writer.Write(p.toRow())
	}
	log.Printf("wrote %v rows to output %v", len(masterMap), outputFile)
}

func collectStatsFromSeason(m map[string]*playerData, season *Season) {
}
