package managers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hm2/check"
	"hm2/config"
	"hm2/convert"
	"hm2/get"
	"hm2/report"
	"hm2/result"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	Surname    string                `json:"surname"`
	NickName   string                `json:"nickname"`
	Sex        bool                  `json:"sex"`
	Country    string                `json:"country"`
	City       string                `json:"city"`
	Birth      string                `json:"birth"`
	Mail       string                `json:"mail"`
	Img        string                `json:"img"`
	LastOnline string                `json:"last_online"`
	Created    string                `json:"created"`
	Cash       float64               `json:"cash"`
	Rating     float64               `json:"rating"`
	Teams      []ProfileManagerTeams `json:"teams"`
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
	Stats []ProfileManagerStatsRecord `json:"stats"`
}

type ProfileManagerStatsRecord struct {
	DateStart string  `json:"datestart"`
	DateEnd   *string `json:"dateend"`
	Team      string  `json:"team"`
	TeamID    int     `json:"teamid"`
	GP        int     `json:"gp"`
	W         int     `json:"w"`
	WO        int     `json:"wo"`
	WS        int     `json:"ws"`
	LS        int     `json:"ls"`
	LO        int     `json:"lo"`
	L         int     `json:"l"`
	TrophyNum int     `json:"trophynum"`
}

type Date struct {
	Day   int
	Month int
	Year  int
}

type ManagerTeams struct {
	Team1      int `db:"team1"`
	Team2      int `db:"team2"`
	Team3      int `db:"team3"`
	NumOfTeams int `db:"team_num"`
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, true)
	if !user.Authenticated {
		return
	}
	res = GetProfile(r, user.ID, user)
	result.ReturnJSON(w, &res)
}

func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, false)
	vars := mux.Vars(r)
	id := vars["id"]
	ID, _ := strconv.Atoi(id)
	res = GetProfile(r, ID, user)
	result.ReturnJSON(w, &res)
}

func ProfileStatsHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, true)
	if !user.Authenticated {
		return
	}
	keys := r.URL.Query()
	mes, page, limit := check.Paginator(keys)
	if mes != `` {
		res = result.SetErrorResult(mes)
		result.ReturnJSON(w, &res)
		return
	}
	res = GetProfileStats(r, user.ID, page, limit)
	result.ReturnJSON(w, &res)
}

func GetProfileStatsHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	IsLogin(w, r, false)
	vars := mux.Vars(r)
	id := vars["id"]
	ID, err := strconv.Atoi(id)
	if err != nil {
		res = result.SetErrorResult(`Неверный параметр id`)
		result.ReturnJSON(w, &res)
		return
	}
	keys := r.URL.Query()
	mes, page, limit := check.Paginator(keys)
	if mes != `` {
		res = result.SetErrorResult(mes)
		result.ReturnJSON(w, &res)
		return
	}
	res = GetProfileStats(r, ID, page, limit)
	result.ReturnJSON(w, &res)
}

func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, true)
	if !user.Authenticated {
		return
	}
	res = EditProfile(r, user.ID)
	result.ReturnJSON(w, &res)
}

func EditProfileConfirmHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, true)
	if !user.Authenticated {
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

func ChangeCurrentTeamHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, true)
	if !user.Authenticated {
		return
	}
	keys := r.URL.Query()
	var TeamNum int
	var err error
	if len(keys[`num`]) > 0 {
		TeamNum, err = strconv.Atoi(keys[`num`][0])
		if err != nil {
			res = result.SetErrorResult(`Неверный параметр номера команды(не число)`)
			result.ReturnJSON(w, &res)
			return
		}
	} else {
		res = result.SetErrorResult(`Необходим параметр num`)
		result.ReturnJSON(w, &res)
		return
	}
	res = ChangeCurrentTeam(r, TeamNum, user)
	result.ReturnJSON(w, &res)
}

func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, true)
	if !user.Authenticated {
		return
	}
	res = UploadImage(r, user.ID)
	result.ReturnJSON(w, &res)
}

func DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, true)
	if !user.Authenticated {
		return
	}
	res = DeleteImage(r, user.ID)
	result.ReturnJSON(w, &res)
}

func GetProfile(r *http.Request, ID int, user config.User) (res result.ResultInfo) {
	var data ProfileManager
	db := config.ConnectDB()
	ctx := r.Context()
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
	err := db.QueryRowContext(ctx, query, ID).Scan(&data.NickName, &data.Mail, &created, &last_online)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, ID)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	data.Created = created.Format("02.01.2006 15:04")
	if last_online.Add(5 * time.Minute).After(time.Now()) {
		data.LastOnline = "online"
	} else {
		data.LastOnline = last_online.Format("02.01.2006 15:04")
	}
	var d Date
	query = `SELECT name, surname, sex, country, city, birthd, birthm, birthy, img, cash from managers.data WHERE id=$1`
	err = db.QueryRowContext(ctx, query, ID).Scan(&data.Name, &data.Surname, &data.Sex, &data.Country, &data.City, &d.Day, &d.Month, &d.Year, &data.Img, &data.Cash)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, ID)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	data.Birth = DateToString(d)
	data.Rating = 234
	data.Teams, err = FillTeamsTest(ctx, ID)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, ID)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	res.Done = true
	res.Items = data
	return res
}

func GetProfileStats(r *http.Request, ID int, Page int, Limit int) (res result.ResultInfo) {
	var data ProfileManagerStats
	db := config.ConnectDB()
	ctx := r.Context()
	var p result.Paginator
	p.Limit = Limit
	p.Page = Page
	IsProfileExist := CheckProfileExist(r, ID)
	if !IsProfileExist {
		res = result.SetErrorResult(`Данного профиля не существует`)
		return
	}
	var pmr ProfileManagerStatsRecord
	query := `SELECT COUNT(*) FROM managers.history WHERE manager_id = $1`
	params := []interface{}{ID}
	err := db.QueryRowContext(ctx, query, params...).Scan(&p.Total)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, ID)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	if p.Total%p.Limit == 0 {
		p.CountPage = p.Total / p.Limit
	} else {
		p.CountPage = (p.Total / p.Limit) + 1
	}
	p.Offset = (p.Page - 1) * p.Limit
	query = `SELECT date_start, date_finish, team_name, G, W, WO, WS, LS, LO, L, trophies FROM managers.history WHERE manager_id = $1 ORDER BY date_start DESC LIMIT $2 OFFSET $3`
	params = []interface{}{ID, p.Limit, p.Offset}
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, ID)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&pmr.DateStart, &pmr.DateEnd, &pmr.Team, &pmr.GP, &pmr.W, &pmr.WO, &pmr.WS, &pmr.LS, &pmr.LO, &pmr.L, &pmr.TrophyNum)
		if err != nil {
			switch {
			case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
				res = result.SetErrorResult(report.CtxError)
			default:
				report.ErrorSQLServer(r, err, query, ID)
				res = result.SetErrorResult(report.UnknownError)
			}
			return
		}
		pmr.TeamID, err = get.TeamIDByTeamName(ctx, pmr.Team)
		if err != nil {
			switch {
			case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
				res = result.SetErrorResult(report.CtxError)
			default:
				report.ErrorSQLServer(r, err, query, ID)
				res = result.SetErrorResult(report.UnknownError)
			}
			return
		}
		data.Stats = append(data.Stats, pmr)
	}
	res.Done = true
	res.Items = data
	res.Paginator = &p
	return
}

func EditProfile(r *http.Request, ID int) (res result.ResultInfo) {
	data := []ProfileManagerEdit{}
	db := config.ConnectDB()
	ctx := r.Context()
	IsProfileExist := CheckProfileExist(r, ID)
	if !IsProfileExist {
		res = result.SetErrorResult(`Данного профиля не существует`)
		return
	}
	query := `SELECT name, surname, sex, country, city, birthd, birthm, birthy, img from managers.data WHERE id=$1`
	err := db.SelectContext(ctx, &data, query, ID)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, ID)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	for i := 0; i < len(data); i++ {
		data[i].ID = ID
	}
	res.Done = true
	res.Items = data
	return res
}

func UploadImage(r *http.Request, ID int) (res result.ResultInfo) {
	file, _, err := r.FormFile("file")
	if err != nil {
		res = result.SetErrorResult(`Ошибка получения файла`)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		res = result.SetErrorResult(`Ошибка чтения файла`)
		return
	}
	FullFileName := fmt.Sprintf("public/profile/%d.%s", ID, `jpg`)
	FileOnDisk, err := os.Create(FullFileName)
	if err != nil {
		res = result.SetErrorResult(`Ошибка создания файла на диске`)
		return
	}
	defer FileOnDisk.Close()
	_, err = FileOnDisk.Write(fileBytes)
	if err != nil {
		res = result.SetErrorResult(`Ошибка записи в файл`)
		return
	}
	db := config.ConnectDB()
	query := `UPDATE managers.data SET img = $1 WHERE id = $2`
	params := []interface{}{ID, ID}
	_, err = db.Exec(query, params...)
	if err != nil {
		res = result.SetErrorResult(`Ошибка обновления базы данных`)
		return
	}
	res.Done = true
	res.Items = map[string]interface{}{"Path": fmt.Sprintf("public/profile/%d.jpg", ID)}
	return
}

func DeleteImage(r *http.Request, ID int) (res result.ResultInfo) {
	db := config.ConnectDB()
	query := `UPDATE managers.data SET img = '-1' WHERE id = $1`
	params := []interface{}{ID}
	_, err := db.Exec(query, params...)
	if err != nil {
		res = result.SetErrorResult(`Ошибка обновления базы данных`)
		return
	}
	os.Remove(fmt.Sprintf(`public/profile/%d.jpg`, ID))
	res.Done = true
	return
}

func EditProfileConfirm(r *http.Request, data ProfileManagerEdit) (res result.ResultInfo) {
	db := config.ConnectDB()
	ctx := r.Context()
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
	_, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, params...)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	res.Done = true
	res.Items = data.ID
	return res
}

func FillTeamsTest(ctx context.Context, IDManager int) (res []ProfileManagerTeams, err error) {
	db := config.ConnectDB()
	var p ProfileManagerTeams
	TeamsID, err := get.TeamsByManager(ctx, IDManager)
	if err != nil {
		return nil, err
	}
	for i := range TeamsID {
		query := `SELECT name, country FROM list.team_list WHERE id = $1`
		params := []interface{}{TeamsID[i]}
		err = db.QueryRowContext(ctx, query, params...).Scan(&p.Team, &p.Country)
		if err != nil {
			return nil, err
		}
		p.TeamID = TeamsID[i]
		p.CountryID, err = convert.NationToInt(p.Country)
		if err != nil {
			return nil, err
		}
		p.Division = "testdiv" + strconv.Itoa(i)
		p.DivisionID = i
		res = append(res, p)
	}
	return res, nil
}

func ChangeCurrentTeam(r *http.Request, NewNum int, user config.User) (res result.ResultInfo) {
	db := config.ConnectDB()
	ctx := r.Context()
	var TeamNum int
	if NewNum > 3 || NewNum < 1 {
		res = result.SetErrorResult(`Номер команды для смены должен быть от 1 до 3`)
		return
	}
	query := `SELECT team_num from list.manager_list WHERE id = $1`
	params := []interface{}{user.ID}
	err := db.QueryRowContext(ctx, query, params...).Scan(&TeamNum)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, params...)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	if TeamNum < NewNum {
		res = result.SetErrorResult(`У пользователя меньше команд, невозможно сменить команду`)
		return
	}
	query = `UPDATE list.manager_list SET cur_team = $1 WHERE id = $2`
	params = []interface{}{NewNum, user.ID}
	_, err = db.ExecContext(ctx, query, params...)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, params...)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	CurTeamID, err := get.CurrentTeamByManager(ctx, user.ID)
	if err != nil {
		res = result.SetErrorResult(err.Error())
		return
	}
	res.Done = true
	res.Items = CurTeamID
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
