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
	"sync"
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

	NUM_COLS = 11
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

type playerData struct {
	Name                                   string
	TotalMakes, TotalChances               int
	ColdMakes, ColdChances                 int
	HotMakes, HotChances                   int
	TotalPct, ColdPct, HotPct, HotColdDiff float64
}

func (pd *playerData) toRow() []string {
	if pd.TotalChances == 0 {
		pd.TotalPct = 0
		pd.ColdPct = 0
		pd.HotPct = 0
	} else if pd.HotChances == 0 {
		pd.TotalPct = float64(pd.TotalMakes) / float64(pd.TotalChances)
		pd.ColdPct = float64(pd.ColdMakes) / float64(pd.ColdChances)
	} else {
		pd.TotalPct = float64(pd.TotalMakes) / float64(pd.TotalChances)
		pd.ColdPct = float64(pd.ColdMakes) / float64(pd.ColdChances)
		pd.HotPct = float64(pd.HotMakes) / float64(pd.HotChances)
	}
	pd.HotColdDiff = pd.HotPct - pd.ColdPct

	row := []string{
		pd.Name,
		fmt.Sprintf("%v", pd.TotalMakes),
		fmt.Sprintf("%v", pd.TotalChances),
		fmt.Sprintf("%v", pd.TotalPct),
		fmt.Sprintf("%v", pd.ColdMakes),
		fmt.Sprintf("%v", pd.ColdChances),
		fmt.Sprintf("%v", pd.ColdPct),
		fmt.Sprintf("%v", pd.HotMakes),
		fmt.Sprintf("%v", pd.HotChances),
		fmt.Sprintf("%v", pd.HotPct),
		fmt.Sprintf("%v", pd.HotColdDiff),
	}
	return row
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

func RunPointsReport(season *Season, player string) []string {

	games := 0
	points := 0

	for _, game := range season.Games {
		inGame := false
		gamePoints := 0

		for _, event := range game {
			if strings.Contains(event.Player, player) {
				inGame = true
				if event.Is3PM() {
					gamePoints += 3
				} else if event.IsFGM() {
					gamePoints += 2
				} else if event.IsFTM() {
					gamePoints += 1
				}
			}
		}

		if inGame {
			games += 1
			points += gamePoints
		}
	}

	var pointsPerGame float64
	if points == 0 {
		pointsPerGame = 0
	} else {
		pointsPerGame = float64(points) / float64(games)
	}

	return []string{
		player,
		fmt.Sprintf("%v", games),
		fmt.Sprintf("%v", points),
		fmt.Sprintf("%v", pointsPerGame),
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
	//for _, p := range players {
	//	fmt.Println(p)
	//}
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
				//log.Printf("Error on NewGameEventFromRow: %v. Skipping row: '%v'", err, row)
				continue
			}
			game = append(game, gd)
		}
		season.Games[file] = game
	}
	return season, nil
}

var outputFile string
var seasonIds []string = []string{
	//"1998_99",
	//"1999_00",
	//"2000_01",
	//"2001_02",
	//"2002_03",
	//"2003_04",
	//"2004_05",
	//"2005_06",
	//"2006_07",
	//"2007_08",
	//"2008_09",
	//"2009_10",
	//"2010_11",
	//"2011_12",
	//"2012_13",
	//"2013_14",
	//"2014_15",
	//"2015_16",
	"2016_17",
}

type WorkQueue struct {
	playerChannel chan string
	wg            *sync.WaitGroup
}

func InitializeWorkQueue(m map[string]*playerData, season *Season, numParallelWorkers int) *WorkQueue {
	wq := &WorkQueue{
		playerChannel: make(chan string, numParallelWorkers),
		wg:            new(sync.WaitGroup),
	}

	for i := 0; i < numParallelWorkers; i++ {
		wq.wg.Add(1)
		tmpi := i
		go func() {
			defer wq.wg.Done()
			for player := range wq.playerChannel {
				log.Printf("Doing the thing from worker #%v", tmpi)
				collectStatsFromSeasonForPlayer(m, season, player)
			}
		}()
	}
	return wq
}

func main() {

	flag.StringVar(&outputFile, "o", "", "output file")
	//flag.StringVar(&sid, "s", "", "season")
	flag.Parse()
	if outputFile == "" {
		log.Fatalf("Must supply -o")
	}
	//if sid == "" {
	//	log.Fatalf("Must supply -s")
	//}

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

	for _, s := range seasons {
		collectStatsFromSeason(masterMap, s)
	}

	for _, p := range masterMap {
		log.Printf("YOYOYO: %v", p.toRow())
	}

	for _, p := range masterMap {
		writer.Write(p.toRow())
	}
	log.Printf("wrote %v rows to output - %v", len(masterMap))
}

func CreatePointsReport(season *Season) [][]string {
	rows := make([][]string, 0, len(season.Players)+1)
	rows = append(rows, []string{"PLAYER", "GAMES", "POINTS", "PPG"})
	for _, player := range season.Players {
		row := RunPointsReport(season, player)
		rows = append(rows, row)
	}
	return rows
}

func collectStatsFromSeasonForPlayer(m map[string]*playerData, season *Season, player string) {
	log.Printf("collecting stats for %v", player)

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
		TotalMakes:   totalMakes,
		TotalChances: totalChances,
		ColdMakes:    coldMakes,
		ColdChances:  coldChances,
		HotMakes:     hotMakes,
		HotChances:   hotChances,
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

func collectStatsFromSeason(m map[string]*playerData, season *Season) {
	for _, player := range season.Players {
		collectStatsFromSeasonForPlayer(m, season, player)
	}
}
