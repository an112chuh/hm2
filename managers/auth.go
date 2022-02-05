package managers

import (
	"bytes"
	"encoding/json"
	"hm2/config"
	"hm2/report"
	"hm2/result"
	"io"
	"net/http"
	"time"
)

type ManagerLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, "cookie-name")
	if err != nil {
		report.ErrorServer(r, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var data ManagerLogin
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	err = json.Unmarshal(b, &data)
	if err != nil {
		report.ErrorServer(r, err)
	}
	res, ID := LoginManager(r, data)
	if res.Done {
		user := &config.User{
			Username:      data.Login,
			ID:            ID,
			Authenticated: true,
		}
		session.Values["user"] = user
		err = session.Save(r, w)
		if err != nil {
			report.ErrorServer(r, err)
			res = result.SetErrorResult(`Внутренняя ошибка`)
		}
		SetOnline(*user)
	}
	result.ReturnJSON(w, &res)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, "cookie-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["user"] = config.User{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var res result.ResultInfo
	res.Done = true
	result.ReturnJSON(w, &res)
}

func LoginManager(r *http.Request, data ManagerLogin) (res result.ResultInfo, ID int) {
	res = FindLogin(r, data)
	if res.Done {
		res, ID = CheckPassword(r, data)
	}
	return res, ID
}

func FindLogin(r *http.Request, data ManagerLogin) (res result.ResultInfo) {
	db := config.ConnectDB()
	var LoginExist bool
	query := `SELECT EXISTS (SELECT 1 FROM list.manager_list WHERE login = $1)`
	err := db.QueryRow(query, data.Login).Scan(&LoginExist)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return
	}
	if !LoginExist {
		res = result.SetErrorResult(`Неверные логин или пароль`)
		return
	}
	res.Done = true
	return res
}

func CheckPassword(r *http.Request, data ManagerLogin) (res result.ResultInfo, ID int) {
	db := config.ConnectDB()
	Hash := HashCreation(data.Password)
	var HashFromDB uint32
	query := "SELECT hash, id FROM list.manager_list WHERE login = $1"
	err := db.QueryRow(query, data.Login).Scan(&HashFromDB, &ID)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return
	}
	if Hash == HashFromDB {
		res.Done = true
		res.Items = ID
	} else {
		res = result.SetErrorResult(`Неверные логин или пароль`)
		return
	}
	return res, ID
}

func SetOnline(user config.User) {
	db := config.ConnectDB()
	t := time.Now()
	query := `UPDATE list.manager_list SET last_online = $1 WHERE id = $2`
	params := []interface{}{t, user.ID}
	_, err := db.Exec(query, params...)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return
	}
}
