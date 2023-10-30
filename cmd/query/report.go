package query

import (
	"fmt"

	"github.com/aldehir/ut2u/pkg/encoding/ue2"
	"github.com/aldehir/ut2u/pkg/query"
)

type Formatter interface {
	Report(rpt Server) error
	Flush() error
}

type Server struct {
	Address   string `json:"address"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	QueryPort int    `json:"query_port"`

	Info    ServerInfo   `json:"info"`
	Rules   []ServerRule `json:"rules"`
	Players []Player     `json:"players"`
	Teams   []Team       `json:"teams"`

	Status struct {
		Success bool   `json:"success"`
		Message string `json:"msg"`
	} `json:"status"`
}

type ServerInfo struct {
	Name       string       `json:"name"`
	NameColors []ColorPoint `json:"name_colors"`

	Map       string       `json:"map"`
	MapColors []ColorPoint `json:"map_colors"`

	GameType       string       `json:"gametype"`
	GameTypeColors []ColorPoint `json:"gametype_colors"`

	Players struct {
		Current int `json:"current"`
		Max     int `json:"max"`
	} `json:"player_count"`

	SkillLevel string `json:"skill_level"`
}

type ServerRule struct {
	Key         string       `json:"key"`
	KeyColors   []ColorPoint `json:"key_colors"`
	Value       string       `json:"value"`
	ValueColors []ColorPoint `json:"value_colors"`
}

type ColorPoint struct {
	Index int   `json:"index"`
	Color Color `json:"color"`
}

type Color struct {
	Red   int    `json:"red"`
	Green int    `json:"green"`
	Blue  int    `json:"blue"`
	Hex   string `json:"hex"`
}

type Player struct {
	Index      int          `json:"index"`
	Team       int          `json:"team"`
	Spectator  bool         `json:"spectator"`
	Name       string       `json:"name"`
	NameColors []ColorPoint `json:"name_colors"`
	Ping       int          `json:"ping"`
	Score      int          `json:"score"`
	StatsID    int          `json:"stats_id"`
}

type Team struct {
	Index      int          `json:"index"`
	Name       string       `json:"name"`
	NameColors []ColorPoint `json:"name_colors"`
	Score      int          `json:"score"`
}

func CreateServerInfo(info query.ServerInfo) ServerInfo {
	var result ServerInfo

	result.Name = info.ServerName.Value
	result.NameColors = CreateColorPoints(info.ServerName.ColorPoints)
	result.Map = info.MapName.Value
	result.MapColors = CreateColorPoints(info.MapName.ColorPoints)
	result.GameType = info.GameType.Value
	result.GameTypeColors = CreateColorPoints(info.GameType.ColorPoints)
	result.Players.Current = int(info.CurrentPlayers)
	result.Players.Max = int(info.MaxPlayers)
	result.SkillLevel = info.SkillLevel

	return result
}

func CreateRules(rules []query.KeyValuePair) []ServerRule {
	result := make([]ServerRule, len(rules))

	for i := 0; i < len(rules); i++ {
		result[i].Key = rules[i].Key.Value
		result[i].KeyColors = CreateColorPoints(rules[i].Key.ColorPoints)
		result[i].Value = rules[i].Value.Value
		result[i].ValueColors = CreateColorPoints(rules[i].Value.ColorPoints)
	}

	return result
}

func CreatePlayersAndTeams(players []query.Player, count int) ([]Player, []Team) {
	p := make([]Player, 0, 16)
	t := make([]Team, 0, 4)

	var i int

	for i = 0; i < len(players); i++ {
		var player Player

		player.Index = int(players[i].Num)
		player.Name = players[i].Name.Value
		player.NameColors = CreateColorPoints(players[i].Name.ColorPoints)
		player.Score = int(players[i].Score)
		player.Ping = int(players[i].Ping)

		// Team information is encoded in StatsID at bits 30 and 29
		player.StatsID = int(players[i].StatsID & (0x1fffffff))

		player.Team = -1
		if players[i].StatsID&(1<<29) != 0 {
			player.Team = 0
		} else if players[i].StatsID&(1<<30) != 0 {
			player.Team = 1
		}

		// Spectators are shown above player count, however some custom game
		// types also encode teams as players at the end. We want to record
		// those with a ping as spectators and then process teams separately
		if i >= count {
			if player.Ping == 0 {
				// Hit the team section, process in next loop
				break
			} else {
				// Assign as a spectator
				player.Team = -1
				player.Spectator = true
			}
		}

		p = append(p, player)
	}

	// Some custom game types encode team score information as the players
	// over playerCount
	for ; i < len(players); i++ {
		var team Team

		team.Index = len(t)
		team.Name = players[i].Name.Value
		team.NameColors = CreateColorPoints(players[i].Name.ColorPoints)
		team.Score = int(players[i].Score)

		t = append(t, team)
	}

	return p, t
}

func CreateColorPoints(colors []ue2.ColorPoint) []ColorPoint {
	points := make([]ColorPoint, len(colors))

	for i := 0; i < len(colors); i++ {
		points[i].Index = colors[i].At

		r, g, b, _ := colors[i].Color.RGBA()

		// Reduce range to [0, 255]
		r = r >> 8
		g = g >> 8
		b = b >> 8

		points[i].Color.Red = int(r)
		points[i].Color.Green = int(g)
		points[i].Color.Blue = int(b)
		points[i].Color.Hex = fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}

	return points
}
