package players

import (
	"database/sql"
	"fmt"
	"hm2/constants"
	"hm2/convert"
	"hm2/report"
	"io"
	"math/rand"
	"net/http"
	"os"
	s "strings"
)

type Nation struct {
	ID        int
	ShortName string
	NameRUS   string
}

type Player struct {
	Id        int       `json:"id"`
	TeamID    int       `json:"team_id"`
	TeamName  string    `json:"team_name"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Age       int       `json:"age"`
	Nat       int       `json:"nat"`
	NatString string    `json:"nat_string"`
	Price     int       `json:"price"`
	Pos       int       `json:"pos"`
	PosString string    `json:"pos_string"`
	IsGK      bool      `json:"is_gk"`
	Str       int       `json:"str"`
	Readyness int       `json:"readyness"`
	Morale    int       `json:"morale"`
	Tireness  int       `json:"tireness"`
	Style     int       `json:"style"`
	Skills    *Skills   `json:"skills,omitempty"`
	GKSkills  *GKSkills `json:"gk_skills,omitempty"`
	History   []History `json:"history"`
}

type Skills struct {
	Speed         int    `json:"speed"`
	Skating       int    `json:"skating"`
	SlapShot      int    `json:"slap_shot"`
	WristShot     int    `json:"wrist_shot"`
	Tackling      int    `json:"tackling"`
	Blocking      int    `json:"blocking"`
	Passing       int    `json:"passing"`
	Vision        int    `json:"vision"`
	Agressiveness int    `json:"agressiveness"`
	Resistance    int    `json:"resistance"`
	Faceoff       int    `json:"faceoff"`
	Hand          string `json:"hand"`
}

type History struct {
	Team   string  `json:"team"`
	GP     int     `json:"GP"`
	G      int     `json:"G"`
	A      int     `json:"A"`
	P      int     `json:"P"`
	PIM    int     `json:"PIM"`
	PM     int     `json:"PM"`
	SOG    int     `json:"SOG"`
	SOfG   int     `json:"SOfG"`
	Rating float64 `json:"rating"`
}

type GKSkills struct {
	StickHandle    int `json:"stick_handle"`
	GloveHandle    int `json:"glove_handle"`
	RicochetContrl int `json:"ricochet_contrl"`
	FiveHole       int `json:"five_hole"`
	Passing        int `json:"passing"`
	Reaction       int `json:"reaction"`
}

type AgeWithStr struct {
	Age int
	Str int
}

func CreatePlayers(r *http.Request, tx *sql.Tx, IDTeam int, NameTeam string, Country string) (*sql.Tx, int, error) {
	var p Player
	str := 0
	var Stats [constants.NumOfSkills]int
	query := `SELECT id, short FROM list.nation_list where name = $1`
	params := []interface{}{Country}
	ctx := r.Context()
	err := tx.QueryRowContext(ctx, query, params...).Scan(&p.Nat, &p.NatString)
	if err != nil {
		report.ErrorSQLServer(r, err, query, params...)
		return tx, str, err
	}
	caps := "txt_files/names/" + s.ToUpper(p.NatString)
	NamesList := ScanNameFromFile(r, caps+"n.txt")
	SurnamesList := ScanNameFromFile(r, caps+"s.txt")
	TeamAges := []AgeWithStr{{16, 42}, {17, 46}, {18, 51}, {19, 54}, {20, 58}, {21, 63}, {22, 65}, {23, 66}, {24, 67}, {25, 69}, {26, 70}, {27, 71}, {28, 72}, {29, 74}, {30, 75}, {31, 76}, {32, 77}, {33, 78}, {34, 80}, {35, 81}, {36, 80}, {37, 74}, {38, 67}, {39, 59},
		{18, 61}, {21, 74}, {24, 81}, {25, 84}, {27, 87}, {30, 90}, {33, 86}}
	for i := 0; i < len(TeamAges); i++ {
		str += TeamAges[i].Str
	}
	rand.Shuffle(len(TeamAges), func(i, j int) {
		TeamAges[i], TeamAges[j] = TeamAges[j], TeamAges[i]
	})
	p.TeamID = IDTeam
	for i := 0; i < 31; i++ {
		p.Name = NamesList[ran(len(NamesList))]
		p.Surname = SurnamesList[ran(len(SurnamesList))]
		p.Age = TeamAges[i].Age
		p.Str = TeamAges[i].Str
		p.Style = ran(9)
		if i < 3 {
			p.Pos = 0
			*p.GKSkills = GenerateGoaliesStats(TeamAges[i])
		} else if i < 8 {
			p.Pos = 1
			Stats = GenerateDefStats(TeamAges[i])
			p.Skills.Hand = "R"
		} else if i < 13 {
			p.Pos = 2
			Stats = GenerateDefStats(TeamAges[i])
			p.Skills.Hand = "L"
		} else if i < 19 {
			p.Pos = 3
			Stats = GenerateWingStats(TeamAges[i])
			p.Skills.Hand = "R"
		} else if i < 25 {
			p.Pos = 4
			Stats = GenerateCenterStats(TeamAges[i])
			x := ran(2)
			if x == 1 {
				p.Skills.Hand = "R"
			} else {
				p.Skills.Hand = "L"
			}
		} else {
			p.Pos = 5
			Stats = GenerateWingStats(TeamAges[i])
			p.Skills.Hand = "L"
		}
		var IDPlayer int
		query = `INSERT into list.players_list (team_id, name, surname, pos, nat, age, str, style, morale, readyness, tireness, price) VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, 100, 100, 0, 1000000) RETURNING id`
		params = []interface{}{p.TeamID, p.Name, p.Surname, p.Pos, p.Nat, p.Age, p.Str, p.Style}
		err = tx.QueryRowContext(ctx, query, params...).Scan(&IDPlayer)
		if err != nil {
			report.ErrorSQLServer(r, err, query, params...)
			return tx, str, err
		}
		if i < 3 {
			query = `INSERT into players.gk_skills (
				player_id, 
				team_id, 
				pos, 
				stick_handle, 
				glove_handle, 
				ricochet_control, 
				fivehole, 
				passing, 
				reaction) VALUES ($1,$2,'GK',$3,$4,$5,$6,$7,$8)`
			params = []interface{}{IDPlayer, IDTeam, p.GKSkills.StickHandle, p.GKSkills.GloveHandle, p.GKSkills.RicochetContrl, p.GKSkills.FiveHole, p.GKSkills.Passing, p.GKSkills.Reaction}
			_, err = tx.ExecContext(ctx, query, params...)
			if err != nil {
				report.ErrorSQLServer(r, err, query, params...)
				return tx, str, err
			}
		}
		if i > 2 {
			p.Skills.Speed = Stats[0]
			p.Skills.Skating = Stats[1]
			p.Skills.SlapShot = Stats[2]
			p.Skills.WristShot = Stats[3]
			p.Skills.Tackling = Stats[4]
			p.Skills.Blocking = Stats[5]
			p.Skills.Passing = Stats[6]
			p.Skills.Vision = Stats[7]
			p.Skills.Agressiveness = Stats[8]
			p.Skills.Resistance = Stats[9]
			p.Skills.Faceoff = Stats[10]
			query = `INSERT into players.skills (player_id, 
				team_id, 
				pos, 
				speed, 
				skating, 
				slap_shot, 
				wrist_shot, 
				tackling, 
				blocking, 
				passing, 
				vision, 
				agressiveness, 
				resistance, 
				faceoff, 
				side) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
			p.PosString = convert.PosToString(p.Pos)
			params := []interface{}{IDPlayer, IDTeam, p.PosString, p.Skills.Speed,
				p.Skills.Skating,
				p.Skills.SlapShot,
				p.Skills.WristShot,
				p.Skills.Tackling,
				p.Skills.Blocking,
				p.Skills.Passing,
				p.Skills.Vision,
				p.Skills.Agressiveness,
				p.Skills.Resistance,
				p.Skills.Faceoff,
				p.Skills.Hand}
			_, err = tx.ExecContext(ctx, query, params...)
			if err != nil {
				report.ErrorSQLServer(r, err, query, params...)
				return tx, str, err
			}
		}
		query = `INSERT into players.history (player_id, team_id, team_name, GP, G, A, P, PIM, PM, SOG, SOfG, rating) VALUES ($1, $2, $3, 0,0,0,0,0,0,0,0,0)`
		params = []interface{}{IDPlayer, IDTeam, NameTeam}
		_, err = tx.ExecContext(ctx, query, params...)
		if err != nil {
			report.ErrorSQLServer(r, err, query, params...)
			return tx, str, err
		}
	}
	return tx, str, err
}

func GenerateGoaliesStats(base AgeWithStr) GKSkills {
	var GKStats [constants.NumOfGKSkills]int
	Str := base.Str * constants.NumOfGKSkills / constants.NumOfSkills
	for i := 0; i < Str; i++ {
		j := ran(constants.NumOfGKSkills)
		GKStats[j] += 1
	}
	for i := 0; i < len(GKStats); i++ {
		if GKStats[i] == 0 {
			GKStats[i]++
		}
	}
	var res GKSkills
	res.FiveHole = GKStats[0]
	res.GloveHandle = GKStats[1]
	res.Passing = GKStats[2]
	res.Reaction = GKStats[3]
	res.RicochetContrl = GKStats[4]
	res.StickHandle = GKStats[5]
	return res
}

func GenerateDefStats(base AgeWithStr) [constants.NumOfSkills]int {
	var DefStats [constants.NumOfSkills]int
	Probs := [constants.NumOfSkills]int{10, 7, 9, 3, 17, 17, 9, 7, 13, 5, 3}
	for CheckStrength(DefStats, base.Str, 1) {
		DefStats[RandomWithStats(Probs)]++
	}
	for i := 0; i < len(DefStats); i++ {
		if DefStats[i] == 0 {
			DefStats[i]++
		}
	}
	return DefStats
}

func GenerateWingStats(base AgeWithStr) [constants.NumOfSkills]int {
	var WingStats [constants.NumOfSkills]int
	Probs := [constants.NumOfSkills]int{17, 17, 4, 11, 4, 8, 8, 11, 5, 12, 3}
	for CheckStrength(WingStats, base.Str, 3) {
		WingStats[RandomWithStats(Probs)]++
	}
	for i := 0; i < len(WingStats); i++ {
		if WingStats[i] == 0 {
			WingStats[i]++
		}
	}
	return WingStats
}

func GenerateCenterStats(base AgeWithStr) [constants.NumOfSkills]int {
	var CenterStats [constants.NumOfSkills]int
	Probs := [constants.NumOfSkills]int{9, 11, 6, 17, 2, 3, 7, 11, 6, 15, 13}
	for CheckStrength(CenterStats, base.Str, 4) {
		CenterStats[RandomWithStats(Probs)]++
	}
	for i := 0; i < len(CenterStats); i++ {
		if CenterStats[i] == 0 {
			CenterStats[i]++
		}
	}
	return CenterStats
}

func ScanNameFromFile(r *http.Request, FileName string) (res []string) {
	var str string
	file, err := os.Open(FileName)
	if err != nil {
		report.ErrorServer(r, err)
		return nil
	}
	defer file.Close()
	for {
		_, err := fmt.Fscanf(file, "%s", &str)
		res = append(res, str)
		if err != nil {
			if err == io.EOF {
				return res
			} else {
				report.ErrorServer(r, err)
				return nil
			}
		}
	}
}

func ran(divider int) int {
	return rand.Intn(divider)
}

func RandomWithStats(p [constants.NumOfSkills]int) int {
	x := ran(100)
	sum := 0
	for i := 0; i < len(p); i++ {
		sum += p[i]
		if x < sum {
			return i
		}
	}
	return len(p)
}

func CheckStrength(Sk [constants.NumOfSkills]int, MaxSkill int, pos int) bool {
	var Sum float64
	if pos == 1 || pos == 2 {
		for i := 0; i < len(Sk); i++ {
			Sum += constants.LDSkills[i] * float64(Sk[i])
		}
	} else if pos == 3 || pos == 5 {
		for i := 0; i < len(Sk); i++ {
			Sum += constants.LWSkills[i] * float64(Sk[i])
		}
	} else if pos == 4 {
		for i := 0; i < len(Sk); i++ {
			Sum += constants.CSkills[i] * float64(Sk[i])
		}
	}
	if int(Sum) < MaxSkill {
		return true
	}
	return false
}
