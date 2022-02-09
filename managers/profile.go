package managers

import (
	"bytes"
	"encoding/json"
	"hm2/config"
	"hm2/report"
	"hm2/result"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type ProfileManagerEdit struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Sex     bool   `json:"sex"`
	Country string `json:"country"`
	City    string `json:"city"`
	BirthD  int    `json:"birthd"`
	BirthM  int    `json:"birthm"`
	BirthY  int    `json:"birthy"`
	Img     string `json:"img"`
}

type ProfileManager struct {
	IsOwned    bool                  `json:"is_owned"`
	ID         int                   `json:"id"`
	Name       string                `json:"name"`
	NickName   string                `json:"nickname"`
	Sex        bool                  `json:"sex"`
	City       string                `json:"city"`
	Birth      string                `json:"birth"`
	Mail       string                `json:"mail"`
	Img        string                `json:"img"`
	LastOnline string                `json:"last_online"`
	Created    string                `json:"created"`
	Teams      []ProfileManagerTeams `json:"teams"`
	Stat       []ProfileManagerStats `json:"stats"`
}

type ProfileManagerTeams struct {
	Country    string `json:"country"`
	CountryID  int    `json:"countryid"`
	Team       string `json:"team"`
	TeamID     int    `json:"teamid"`
	Division   string `json:"division"`
	DivisionID int    `json:"divisionid"`
}

type ProfileManagerStats struct {
	SeasonStart int    `json:"seasonstart"`
	DateStart   string `json:"datestart"`
	SeasonEnd   int    `json:"seasonend"`
	DateEnd     string `json:"dateend"`
	Team        string `json:"team"`
	TeamID      int    `json:"teamid"`
	GP          int    `json:"gp"`
	W           int    `json:"w"`
	L           int    `json:"l"`
	TrophyNum   int    `json:"trophynum"`
}

type Date struct {
	Day   int
	Month int
	Year  int
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {

}

func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	_, user := IsLogin(w, r, false)
	vars := mux.Vars(r)
	id := vars["id"]
	ID, _ := strconv.Atoi(id)
	res = GetProfile(r, ID, user)
	result.ReturnJSON(w, &res)
}

func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	IsLogged, user := IsLogin(w, r, true)
	if !IsLogged {
		return
	}
	res = EditProfile(r, user.ID)
	result.ReturnJSON(w, &res)
}

func EditProfileConfirmHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	IsLogged, user := IsLogin(w, r, true)
	if !IsLogged {
		return
	}
	var data ProfileManagerEdit
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	err := json.Unmarshal(b, &data)
	if err != nil {
		report.ErrorServer(r, err)
	}
	data.ID = user.ID
	res = EditProfileConfirm(r, data)
	result.ReturnJSON(w, &res)
}

func GetProfile(r *http.Request, ID int, user config.User) (res result.ResultInfo) {
	var data ProfileManager
	db := config.ConnectDB()
	IsProfileExist := CheckProfileExist(r, ID)
	if !IsProfileExist {
		res = result.SetErrorResult(`Данного профиля не существует`)
		return
	}
	data.IsOwned = false
	if ID == user.ID && user.Authenticated {
		data.IsOwned = true
	}
	data.ID = ID
	var created, last_online time.Time
	query := `SELECT login, mail, created_at, last_online from list.manager_list where id = $1`
	err := db.QueryRow(query, ID).Scan(&data.NickName, &data.Mail, &created, &last_online)
	if err != nil {
		res = result.SetErrorResult(`Ошибка сервера`)
		report.ErrorServer(r, err)
		return
	}
	data.Created = created.Format("02.01.2006 15:04")
	if last_online.Add(5 * time.Minute).After(time.Now()) {
		data.LastOnline = "online"
	} else {
		data.LastOnline = last_online.Format("02.01.2006 15:04")
	}
	var d Date
	var name, surname, country, city string
	query = `SELECT name, surname, sex, country, city, birthd, birthm, birthy, img from managers.data WHERE id=$1`
	err = db.QueryRow(query, ID).Scan(&name, &surname, &data.Sex, &country, &city, &d.Day, &d.Month, &d.Year, &data.Img)
	if err != nil {
		res = result.SetErrorResult(`Ошибка сервера`)
		report.ErrorServer(r, err)
		return
	}
	data.Name = name + " " + surname
	data.City = country + ", " + city
	data.Birth = DateToString(d)
	data.Teams = FillTeamsTest()
	data.Stat = FillStatsTest()
	res.Done = true
	res.Items = data
	return res
}

func EditProfile(r *http.Request, ID int) (res result.ResultInfo) {
	data := []ProfileManagerEdit{}
	db := config.ConnectDB()
	IsProfileExist := CheckProfileExist(r, ID)
	if !IsProfileExist {
		res = result.SetErrorResult(`Данного профиля не существует`)
		return
	}
	query := `SELECT name, surname, sex, country, city, birthd, birthm, birthy, img from managers.data WHERE id=$1`
	err := db.Select(&data, query, ID)
	if err != nil {
		res = result.SetErrorResult(`Ошибка сервера`)
		report.ErrorServer(r, err)
		return
	}
	for i := 0; i < len(data); i++ {
		data[i].ID = ID
	}
	res.Done = true
	res.Items = data
	return res
}

func EditProfileConfirm(r *http.Request, data ProfileManagerEdit) (res result.ResultInfo) {
	db := config.ConnectDB()
	IsProfileExist := CheckProfileExist(r, data.ID)
	if !IsProfileExist {
		res = result.SetErrorResult(`Данного профиля не существует`)
		return
	}
	query := `UPDATE managers.data SET 
	name = $1,
	surname = $2,
	sex = $3,
	country = $4,
	city = $5,
	birthd = $6,
	birthm = $7,
	birthy = $8,
	img = $9
	WHERE id = $10`
	params := []interface{}{data.Name, data.Surname, data.Sex, data.Country, data.City, data.BirthD, data.BirthM, data.BirthY, data.Img, data.ID}
	_, err := db.Exec(query, params...)
	if err != nil {
		res = result.SetErrorResult(``)
		report.ErrorServer(r, err)
		return
	}
	res.Done = true
	res.Items = data.ID
	return res
}

func FillTeamsTest() (res []ProfileManagerTeams) {
	var p ProfileManagerTeams
	for i := 0; i < 3; i++ {
		p.Country = "testctr" + strconv.Itoa(i)
		p.CountryID = i
		p.Division = "testdiv" + strconv.Itoa(i)
		p.DivisionID = i
		p.Team = "testteam" + strconv.Itoa(i)
		p.TeamID = i
		res = append(res, p)
	}
	return res
}

func FillStatsTest() (res []ProfileManagerStats) {
	var p ProfileManagerStats
	for i := 0; i < 3; i++ {
		p.DateStart = "0" + strconv.Itoa(2+i) + ".02.2022"
		p.DateEnd = "0" + strconv.Itoa(5+i) + ".02.2022"
		p.GP = i
		p.L = i - 1
		p.SeasonEnd = i + 4
		p.SeasonStart = i + 2
		p.Team = "testteam" + strconv.Itoa(i)
		p.TeamID = i
		p.TrophyNum = i
		p.W = 1
		res = append(res, p)
	}
	return res
}

func CheckProfileExist(r *http.Request, ID int) bool {
	db := config.ConnectDB()
	var exist bool
	query := `SELECT EXISTS(SELECT 1 from list.manager_list where id = $1 and is_active = TRUE)`
	err := db.QueryRow(query, ID).Scan(&exist)
	if err != nil {
		report.ErrorServer(r, err)
	}
	return exist
}

func DateToString(d Date) string {
	res := ""
	if d.Day <= 9 {
		res += "0"
	}
	res = res + strconv.Itoa(d.Day) + "."
	if d.Month <= 9 {
		res += "0"
	}
	res = res + strconv.Itoa(d.Month) + "." + strconv.Itoa(d.Year)
	return res
}
