package managers

import (
	"bytes"
	"encoding/json"
	"hm2/config"
	"hm2/report"
	"hm2/result"
	"io"
	"net/http"
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

type Date struct {
	Day   int
	Month int
	Year  int
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {

}

func IDProfileHandler(w http.ResponseWriter, r *http.Request) {

}

func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	IsLogged, user := IsLogin(w, r)
	if !IsLogged {
		res = result.SetErrorResult(`Пожалуйста, авторизуйтесь для данной операции`)
		return
	}
	res = EditProfile(r, user.ID)
	result.ReturnJSON(w, &res)
}

func EditProfileConfirmHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	IsLogged, user := IsLogin(w, r)
	if !IsLogged {
		res = result.SetErrorResult(`Пожалуйста, авторизуйтесь для данной операции`)
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
		res = result.SetErrorResult(``)
		report.ErrorServer(r, err)
		return
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

func CheckProfileExist(r *http.Request, ID int) bool {
	db := config.ConnectDB()
	var exist bool
	query := `SELECT EXISTS(SELECT 1 from list.manager_list where id = $1)`
	err := db.QueryRow(query, ID).Scan(&exist)
	if err != nil {
		report.ErrorServer(r, err)
	}
	return exist
}
