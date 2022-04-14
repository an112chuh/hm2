package managers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hash/fnv"
	"hm2/config"
	"hm2/report"
	"hm2/result"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

type ManagerReg struct {
	Mail     string `json:"mail,omitempty"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password"`
}

func RegManagerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, "cookie-name")
	if err != nil {
		report.ErrorServer(r, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var data ManagerReg
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	err = json.Unmarshal(b, &data)
	if err != nil {
		report.ErrorServer(r, err)
	}
	res, user := RegManager(r, data)
	if res.Done {
		session.Values["user"] = user
		err = session.Save(r, w)
		if err != nil {
			report.ErrorServer(r, err)
			res = result.SetErrorResult(`Внутренняя ошибка`)
		}
	} else {
		result.ReturnJSON(w, &res)
		return
	}
	config.InitUserLogger(user.ID)
	result.ReturnJSON(w, &res)
}

func DeleteManagerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, "cookie-name")
	if err != nil {
		report.ErrorServer(r, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	User := getUser(session)
	res := DeleteManager(r, User.ID)
	if res.Done {
		session.Values["user"] = config.User{}
		session.Options.MaxAge = -1
		err = session.Save(r, w)
		if err != nil {
			report.ErrorServer(r, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	result.ReturnJSON(w, &res)
}

func EditPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := IsLogin(w, r, true)
	if !user.Authenticated {
		res = result.SetErrorResult("Требуется регистрация")
		result.ReturnJSON(w, &res)
		return
	}
	var data ManagerReg
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	err := json.Unmarshal(b, &data)
	if err != nil {
		report.ErrorServer(r, err)
	}
	res = EditPassword(r, data, user)
	result.ReturnJSON(w, &res)
}

func RegManager(r *http.Request, data ManagerReg) (res result.ResultInfo, user config.User) {
	if len(data.Login) > 30 || len(data.Mail) > 30 {
		res = result.SetErrorResult(`Слишком длинные логин/почта, объём должен быть меньше 30 символов`)
		return
	}
	db := config.ConnectDB()
	var LoginExist bool
	query := "SELECT EXISTS(SELECT 1 FROM list.manager_list WHERE login = $1 AND is_active = TRUE)"
	err := db.QueryRow(query, data.Login).Scan(&LoginExist)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return
	}
	if LoginExist {
		res = result.SetErrorResult(`Пользователь с таким логином уже существует`)
		return
	}
	var MailExist bool
	query = "SELECT EXISTS(SELECT 1 FROM list.manager_list WHERE mail = $1 AND is_active = TRUE)"
	err = db.QueryRow(query, data.Mail).Scan(&MailExist)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return
	}
	if MailExist {
		res = result.SetErrorResult(`Пользователь с такой почтой уже существует`)
		return
	}
	res, ID := CreateManager(r, data)
	if ID < 0 {
		return
	}
	user.ID = ID
	user.Username = data.Login
	user.Authenticated = true
	res.Done = true
	res.Items = user.ID
	return res, user
}

func DeleteManager(r *http.Request, id int) (res result.ResultInfo) {
	db := config.ConnectDB()
	query := `UPDATE list.manager_list SET is_active = FALSE, team1 = 0, team2 = 0, team3 = 0 WHERE id = $1`
	_, err := db.Exec(query, id)
	if err != nil {
		report.ErrorSQLServer(r, err, query, id)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return res
	}
	res.Done = true
	res.Items = id
	return res
}

func CreateManager(r *http.Request, data ManagerReg) (res result.ResultInfo, ID int) {
	db := config.ConnectDB()
	tx, err := db.Beginx()
	if err != nil {
		report.ErrorServer(r, err)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()
	Hash := HashCreation(data.Password)
	t := time.Now()
	query := tx.Rebind(`INSERT INTO list.manager_list (
		login
		, hash
		, mail
		, team1
		, team2
		, team3
		, nationalteam
		, rating
		, rights
		, created_at
		, last_online
		, is_active
		, vip
		, team_num
		, cur_team)
		VALUES ($1, $2, $3, 0, 0, 0, 0, 0, 1, $4, $5, TRUE, 0, 0, 0) RETURNING id`)
	params := []interface{}{data.Login, Hash, data.Mail, t, t}
	err = tx.QueryRow(query, params...).Scan(&ID)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return res, -1
	}
	query = tx.Rebind(`INSERT INTO managers.data (id, name, surname, sex, country, city, birthd, birthm, birthy, img)
		VALUES 
		($1, '', '', true, '', '', -1, -1, -1, '')`)
	_, err = tx.Exec(query, ID)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return res, -1
	}
	if err = tx.Commit(); err != nil {
		report.ErrorServer(r, err)
		result.SetErrorResult(`Внутренняя ошибка`)
		return
	}
	return res, ID
}

func EditPassword(r *http.Request, data ManagerReg, user config.User) (res result.ResultInfo) {
	db := config.ConnectDB()
	ctx := r.Context()
	Hash := HashCreation(data.Password)
	query := `UPDATE list.manager_list SET hash = $1 WHERE id = $2`
	params := []interface{}{Hash, user.ID}
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
	res.Items = user.ID
	return
}

func HashCreation(password string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(password))
	return h.Sum32()
}

func getUser(s *sessions.Session) config.User {
	val := s.Values["user"]
	var user = config.User{}
	user, ok := val.(config.User)
	if !ok {
		return config.User{Authenticated: false}
	}
	return user
}

func IsLogin(w http.ResponseWriter, r *http.Request, IsMessageRequired bool) (user config.User) {
	session, err := config.Store.Get(r, "cookie-name")
	if err != nil {
		report.ErrorServer(r, err)
		return
	}
	user = getUser(session)
	if auth := user.Authenticated; !auth {
		err = session.Save(r, w)
		if err != nil {
			report.ErrorServer(r, err)
			return
		}
	}
	SetOnline(user)
	if !user.Authenticated && IsMessageRequired {
		response := result.SetErrorResult(`Пожалуйста, авторизуйтесь для завершения данной операции`)
		result.ReturnJSON(w, &response)
	}
	return user
}
