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

type ManagerLogin struct {
	Mail           string `json:"mail"`
	Login          string `json:"login"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
}

var store *sessions.CookieStore

func RegManagerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, "cookie-name")
	if err != nil {
		report.ErrorServer(r, err)
		return
	}
	var data ManagerLogin
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

func RegManager(r *http.Request, data ManagerLogin) (res result.ResultInfo, user config.User) {
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
	query := "SELECT EXISTS(SELECT 1 FROM list.manager_list WHERE login = $1)"
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
	query = "SELECT EXISTS(SELECT 1 FROM list.manager_list WHERE mail = $1)"
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

func CreateManager(r *http.Request, data ManagerLogin) (res result.ResultInfo, ID int) {
	db := config.ConnectDB()
	Hash := HashCreation(data.Password)
	t := time.Now()
	query := `INSERT INTO list.manager_list (
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
		, last_online)
		VALUES ($1, $2, $3, 0, 0, 0, 0, 0, 1, $4, $5) RETURNING id`
	params := []interface{}{data.Login, Hash, data.Mail, t, t}
	err := db.QueryRow(query, params...).Scan(&ID)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return res, -1
	}
	return res, ID
}

func HashCreation(password string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(password))
	return h.Sum32()
}
