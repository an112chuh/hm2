package managers

import (
	"bytes"
	"encoding/json"
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
	Mail           string `json:"mail"`
	Login          string `json:"login"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
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

func RegManager(r *http.Request, data ManagerReg) (res result.ResultInfo, user config.User) {
	if data.Password != data.PasswordRepeat {
		res = result.SetErrorResult(`Поля "пароль" и "введите пароль" не совпадают`)
		return
	}
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
		, is_active)
		VALUES ($1, $2, $3, 0, 0, 0, 0, 0, 1, $4, $5, TRUE) RETURNING id`)
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

func IsLogin(w http.ResponseWriter, r *http.Request, IsMessageRequired bool) (res bool, user config.User) {
	res = false
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
	} else {
		res = true
	}
	SetOnline(user)
	if !res && IsMessageRequired {
		response := result.SetErrorResult(`Пожалуйста, авторизуйтесь для завершения данной операции`)
		result.ReturnJSON(w, &response)
	}
	return res, user
}
