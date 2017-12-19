package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	TIME_IDX         int = 7
	PERIOD_IDX       int = 8
	PLAYER_IDX       int = 13
	HOME_DESC_IDX    int = 5
	AWAY_DESC_IDX    int = 32
	QUARTER_TIME_SEC int = 12 * 60
)

type Season struct {
	Id      string
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
		Id:      seasonId,
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
				// this is expected to log an error on the header of every game .csv file
				//log.Printf("Error on NewGameEventFromRow: %v. Skipping row: '%v'", err, row)
				continue
			}
			game = append(game, gd)
		}
		season.Games[file] = game
	}
	return season, nil
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

type playerData struct {
	Name                      string
	TotalMakes, TotalChances  int
	ColdMakes, ColdChances    int
	HotMakes, HotChances      int
	TotalPct, ColdPct, HotPct float64
	HotColdDiff, HotMakeup    float64
}

func (pd *playerData) toRow() []string {
	if pd.TotalChances == 0 {
		pd.TotalPct = 0
		pd.ColdPct = 0
		pd.HotPct = 0
		pd.HotMakeup = 0
	} else if pd.HotChances == 0 {
		pd.TotalPct = (float64(pd.TotalMakes) / float64(pd.TotalChances)) * 100
		pd.ColdPct = (float64(pd.ColdMakes) / float64(pd.ColdChances)) * 100
		pd.HotMakeup = 0
	} else {
		pd.TotalPct = (float64(pd.TotalMakes) / float64(pd.TotalChances)) * 100
		pd.ColdPct = (float64(pd.ColdMakes) / float64(pd.ColdChances)) * 100
		pd.HotPct = (float64(pd.HotMakes) / float64(pd.HotChances)) * 100
		pd.HotMakeup = (float64(pd.HotChances) / float64(pd.TotalChances)) * 100
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
		fmt.Sprintf("%v", pd.HotMakeup),
	}
	return row
}

func getHeaderRow() []string {
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
