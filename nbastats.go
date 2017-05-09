package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	TIME_IDX      int = 7
	PERIOD_IDX    int = 8
	PLAYER_IDX    int = 13
	HOME_DESC_IDX int = 5
	AWAY_DESC_IDX int = 32

	QUARTER_TIME_SEC int = 12 * 60

	HOT_SEC int = 60 * 2

	DATA_DIR string = "data"
)

type Season struct {
	Players []string
	Games   map[string][]*GameEvent
}

type GameEvent struct {
	Player  string
	Desc    string
	TimeSec int
	Period  int
	MinLeft int
	SecLeft int
}

func NewGameEventFromRow(row []string) (*GameEvent, error) {
	var err error
	gd := &GameEvent{
		Player: strings.ToLower(row[PLAYER_IDX]),
		Desc:   fmt.Sprintf("%v%v", strings.ToLower(row[HOME_DESC_IDX]), strings.ToLower(row[AWAY_DESC_IDX])),
	}

	timeSplit := strings.Split(row[TIME_IDX], ":")
	if len(timeSplit) != 2 {
		return nil, fmt.Errorf("Malformed time: '%v'", row[TIME_IDX])
	}

	gd.MinLeft, err = strconv.Atoi(timeSplit[0])
	if err != nil {
		return nil, err
	}

	gd.SecLeft, err = strconv.Atoi(timeSplit[1])
	if err != nil {
		return nil, err
	}

	gd.Period, err = strconv.Atoi(row[PERIOD_IDX])
	if err != nil {
		return nil, err
	}

	periodBaseSec := (gd.Period - 1) * QUARTER_TIME_SEC
	cumSecLeft := (gd.MinLeft * 60) + gd.SecLeft
	gd.TimeSec = QUARTER_TIME_SEC - cumSecLeft + periodBaseSec

	return gd, nil
}

func (gd *GameEvent) String() string {
	return fmt.Sprintf("Q%v %v:%v | %v", gd.Period, gd.MinLeft, gd.SecLeft, gd.Desc)
}

func (gd *GameEvent) IsFGA() bool {
	return !strings.Contains(gd.Desc, "free throw") && (strings.Contains(gd.Desc, "pts") || strings.Contains(gd.Desc, "miss"))
}

func (gd *GameEvent) IsFGM() bool {
	return !strings.Contains(gd.Desc, "free throw") && strings.Contains(gd.Desc, "pts")
}

func (gd *GameEvent) Is3PA() bool {
	return strings.Contains(gd.Desc, "3pt")
}

func (gd *GameEvent) Is3PM() bool {
	return strings.Contains(gd.Desc, "3pt") && !strings.Contains(gd.Desc, "miss")
}

func (gd *GameEvent) IsFTA() bool {
	return strings.Contains(gd.Desc, "free throw")
}

func (gd *GameEvent) IsFTM() bool {
	return strings.Contains(gd.Desc, "free throw") && strings.Contains(gd.Desc, "pts")
}

func RunHotColdReport(season *Season, player string) []string {

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

	var totalPct, hotPct, coldPct, hotColdDiff float64
	if totalChances == 0 {
		totalPct = 0
		coldPct = 0
		hotPct = 0
		hotColdDiff = 0
	} else if hotChances == 0 {
		totalPct = 100 * (float64(totalMakes) / float64(totalChances))
		coldPct = 100 * (float64(coldMakes) / float64(coldChances))
		hotPct = 0
		hotColdDiff = 0
	} else {
		totalPct = 100 * (float64(totalMakes) / float64(totalChances))
		coldPct = 100 * (float64(coldMakes) / float64(coldChances))
		hotPct = 100 * (float64(hotMakes) / float64(hotChances))
		hotColdDiff = hotPct - coldPct
	}

	return []string{
		player,
		fmt.Sprintf("%v", totalMakes),
		fmt.Sprintf("%v", totalChances),
		fmt.Sprintf("%v", totalPct),

		fmt.Sprintf("%v", coldMakes),
		fmt.Sprintf("%v", coldChances),
		fmt.Sprintf("%v", coldPct),

		fmt.Sprintf("%v", hotMakes),
		fmt.Sprintf("%v", hotChances),
		fmt.Sprintf("%v", hotPct),

		fmt.Sprintf("%v", hotColdDiff),
	}
}

func ReadPlayersFile(playersFile string) ([]string, error) {
	f, err := os.Open(playersFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	players := make([]string, 0)
	reader := bufio.NewReader(f)
	for {
		player, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		players = append(players, strings.ToLower(strings.Trim(player, " \n")))
	}
	return players, nil
}

func NewSeason(seasonId string) (*Season, error) {

	fnGlob := path.Join(DATA_DIR, seasonId, "*.csv")
	playersFile := path.Join(DATA_DIR, seasonId, "players.dat")

	files, err := filepath.Glob(fnGlob)
	if err != nil {
		return nil, err
	}

	players, err := ReadPlayersFile(playersFile)
	if err != nil {
		return nil, err
	}

	season := &Season{
		Games:   make(map[string][]*GameEvent, len(files)),
		Players: players,
	}

	for _, file := range files {

		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		reader := csv.NewReader(f)
		rows, err := reader.ReadAll()
		if err != nil {
			return nil, err
		}

		game := make([]*GameEvent, 0, len(rows))
		for _, row := range rows {
			gd, err := NewGameEventFromRow(row)
			if err != nil {
				//log.Printf("Error on NewGameEventFromRow: %v. Skipping row.", err)
				continue
			}
			game = append(game, gd)
		}
		season.Games[file] = game
	}
	return season, nil
}

var outputFile, sid string

func main() {

	flag.StringVar(&outputFile, "o", "", "output file")
	flag.StringVar(&sid, "s", "", "season")
	flag.Parse()
	if outputFile == "" {
		log.Fatalf("Must supply -o")
	}
	if sid == "" {
		log.Fatalf("Must supply -s")
	}

	//seasons := make([]*Season, len(seasonIds))

	//for i, sid := range seasonIds {

	//	season, err := NewSeason(sid)
	//	if err != nil {
	//		log.Fatalf("Error creating new season %v: %v", sid, err)
	//	}
	//	log.Printf("built season %s", sid)

	//	seasons[i] = season
	//}

	t := time.Now()

	season, err := NewSeason(sid)
	if err != nil {
		log.Fatalf("error creating new season %v: %v", sid, err)
	}
	log.Printf("built season %v - %v", sid, time.Since(t))

	of, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer of.Close()
	writer := csv.NewWriter(of)

	t = time.Now()

	rows := CreateHotColdReport(season)
	log.Printf("ran report - %v", time.Since(t))

	t = time.Now()

	for _, row := range rows {
		writer.Write(row)
	}
	log.Printf("wrote to output - %v", time.Since(t))
}

// row := RunHotColdReport(season, player)

func CreateHotColdReport(season *Season) [][]string {

	rows := make([][]string, 0, len(season.Players)+1)

	rows = append(rows, []string{"PLAYER", "3PM", "3PA", "3P%", "C_3PM", "C_3PA", "C_3P%", "H_3PM", "H_3PA", "H_3P%", "H_C_DIFF"})

	for _, player := range season.Players {

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

		var totalPct, hotPct, coldPct, hotColdDiff float64

		if totalChances == 0 {
			totalPct = 0
			coldPct = 0
			hotPct = 0
			hotColdDiff = 0
		} else if hotChances == 0 {
			totalPct = 100 * (float64(totalMakes) / float64(totalChances))
			coldPct = 100 * (float64(coldMakes) / float64(coldChances))
			hotPct = 0
			hotColdDiff = 0
		} else {
			totalPct = 100 * (float64(totalMakes) / float64(totalChances))
			coldPct = 100 * (float64(coldMakes) / float64(coldChances))
			hotPct = 100 * (float64(hotMakes) / float64(hotChances))
			hotColdDiff = hotPct - coldPct
		}

		rows = append(rows, []string{
			player,
			fmt.Sprintf("%v", totalMakes),
			fmt.Sprintf("%v", totalChances),
			fmt.Sprintf("%v", totalPct),

			fmt.Sprintf("%v", coldMakes),
			fmt.Sprintf("%v", coldChances),
			fmt.Sprintf("%v", coldPct),

			fmt.Sprintf("%v", hotMakes),
			fmt.Sprintf("%v", hotChances),
			fmt.Sprintf("%v", hotPct),

			fmt.Sprintf("%v", hotColdDiff),
		})
	}
	return rows
}
