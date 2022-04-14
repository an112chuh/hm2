package teams

import (
	"errors"
	"fmt"
	"hm2/check"
	"hm2/config"
	"hm2/constants"
	"hm2/convert"
	"hm2/managers"
	"hm2/report"
	"hm2/result"
	"net/http"
	"strconv"
	s "strings"
)

type FreeTeamParams struct {
	Name     string `json:"name"`
	SortType string `json:"sort_type"`
}

type FreeTeamFilter struct {
	Filter       string
	Value        int
	StringValues []string
}

type FreeTeamsData struct {
	NumOfFreeTeams []NumOfFreeTeams `json:"num_of_teams"`
	Items          []FreeTeamsItem  `json:"table"`
}

type NumOfFreeTeams struct {
	Country   string `json:"country"`
	CountryID int    `json:"country_id"`
	Num       int    `json:"num"`
}

type FreeTeamsItem struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Country    string  `json:"country"`
	CountryInt int     `json:"country_int"`
	Division   string  `json:"division"`
	Place      int     `json:"place"`
	Cash       int     `json:"cash"`
	AverageStr float64 `json:"average_str"`
	Str        int     `json:"str"`
	Cost       int     `json:"cost"`
	Price      int     `json:"price"`
	SellType   bool    `json:"auction"`
}

func FreeTeamsListHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := managers.IsLogin(w, r, true)
	if !user.Authenticated {
		res = result.SetErrorResult("Требуется регистрация")
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
	res = FreeTeamsList(r, keys, page, limit)
	result.ReturnJSON(w, &res)
}

func FreeTeamsList(r *http.Request, keys map[string][]string, page int, limit int) (res result.ResultInfo) {
	db := config.ConnectDB()
	ctx := r.Context()
	FilterStrings, res := AddFilters(r, keys)
	var ft FreeTeamsData
	if !res.Done {
		return
	}
	NameExist := false
	if len(keys[`name`]) > 0 {
		NameExist = true
	}
	var p result.Paginator
	p.Page = page
	p.Limit = limit
	p.Offset = (page - 1) * limit
	queryCount := `SELECT COUNT(*) FROM teams.data inner join list.team_list on list.team_list.id = teams.data.team_id where manager_id = -1 `
	queryString := `SELECT teams.data.team_id, teams.data.name, country, cash, avg_str, str, cost, price, is_auction from teams.data inner join list.team_list on list.team_list.id = teams.data.team_id where manager_id = -1 `
	var queryAdd string
	queryParams := []interface{}{}
	if NameExist {
		queryAdd += "and teams.data.name like ? "
		queryParams = append(queryParams, keys[`name`][0]+"%")
	}
	if len(FilterStrings) > 0 {
		if len(FilterStrings[len(FilterStrings)-1].StringValues) == 0 {
			for i := 0; i < len(FilterStrings); i++ {
				queryAdd += "and " + FilterStrings[i].Filter + "? "
				queryParams = append(queryParams, FilterStrings[i].Value)
			}
		} else {
			for i := 0; i < len(FilterStrings)-1; i++ {
				queryAdd += "and " + FilterStrings[i].Filter + "? "
				queryParams = append(queryParams, FilterStrings[i].Value)
			}
			LastIndex := len(FilterStrings) - 1
			queryAdd += "and "
			for i := 0; i < len(FilterStrings[LastIndex].StringValues); i++ {
				if i == 0 {
					queryAdd += "("
				}
				queryAdd += FilterStrings[LastIndex].Filter + "? "
				queryParams = append(queryParams, FilterStrings[LastIndex].StringValues[i])
				if i != len(FilterStrings[LastIndex].StringValues)-1 {
					queryAdd += "or "
				} else {
					queryAdd += ") "
				}
			}
		}
	}
	queryCount += queryAdd
	queryCount = db.Rebind(queryCount)
	var count int
	err := db.QueryRowContext(ctx, queryCount, queryParams...).Scan(&count)
	if err != nil {
		fmt.Println(queryParams...)
		fmt.Println(queryCount)
		report.ErrorSQLServer(r, err, queryCount, queryParams...)
		res = result.SetErrorResult(report.UnknownError)
		return
	}
	p.Total = count
	if count%limit == 0 {
		p.CountPage = count / limit
	} else {
		p.CountPage = count
	}
	queryString += queryAdd
	SortString, err := SortString(r, keys)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(err.Error())
		return
	}
	queryString += SortString
	query := db.Rebind(queryString)
	rows, err := db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		report.ErrorSQLServer(r, err, query, queryParams...)
		res = result.SetErrorResult(report.UnknownError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var f FreeTeamsItem
		err = rows.Scan(&f.ID, &f.Name, &f.Country, &f.Cash, &f.AverageStr, &f.Str, &f.Cost, &f.Price, &f.SellType)
		if err != nil {
			res = result.SetErrorResult(report.UnknownError)
			report.ErrorServer(r, err)
			return
		}
		f.CountryInt, err = convert.NationToInt(f.Country)
		if err != nil {
			res = result.SetErrorResult(`Запрошенной страны не существует`)
			report.ErrorServer(r, err)
			return
		}
		ft.Items = append(ft.Items, f)
	}
	for i := 0; i < len(constants.CountryList); i++ {
		var n NumOfFreeTeams
		n.Country = constants.CountryList[i]
		query := "SELECT COUNT (*) FROM list.team_list where country = $1"
		params := []interface{}{constants.CountryList[i]}
		err = db.QueryRowContext(ctx, query, params...).Scan(&n.Num)
		if err != nil {
			res = result.SetErrorResult(`Отсутствует страна`)
			report.ErrorSQLServer(r, err, query, params...)
			return
		}
		n.CountryID, err = convert.NationToInt(n.Country)
		if err != nil {
			res = result.SetErrorResult(`Запрошенной страны не существует`)
			report.ErrorServer(r, err)
			return
		}
		ft.NumOfFreeTeams = append(ft.NumOfFreeTeams, n)
	}
	res.Done = true
	res.Items = ft
	res.Paginator = &p
	return
}

func SortString(r *http.Request, keys map[string][]string) (string, error) {
	res := ""
	var err error
	IsSortable := false
	if len(keys[`sort`]) > 0 {
		SortType := keys[`sort`][0]
		switch SortType {
		case "str":
			res += "order by str "
			IsSortable = true
		case "cash":
			res += "order by cash "
			IsSortable = true
		case "cost":
			res += "order by cost "
			IsSortable = true
		case "price":
			res += "order by price "
			IsSortable = true
		case "buy_only":
			res += "and is_auction = false "
		case "auc_only":
			res += "and is_auction = true "
		default:
			err = errors.New("данного параметра сортировки не существует")
			return res, err
		}
	}
	if IsSortable {
		if len(keys[`asc`]) > 0 {
			SortSide := keys[`asc`][0]
			switch SortSide {
			case "true":
				res += "asc"
			case "false":
				res += "desc"
			default:
				err = errors.New("необходимо указание возрастания/убывания")
				return res, err
			}
		} else {
			err = errors.New("необходимо указание возрастания/убывания")
			return res, err
		}
	}
	return res, nil
}

func AddFilters(r *http.Request, keys map[string][]string) (FilterStrings []FreeTeamFilter, res result.ResultInfo) {
	db := config.ConnectDB()
	ctx := r.Context()
	var err error
	var ft FreeTeamFilter
	if len(keys[`price_min`]) > 0 {
		ft.Filter = "price >= "
		ft.Value, err = strconv.Atoi(keys[`price_min`][0])
		if err != nil {
			res = result.SetErrorResult(`Ошибка в запросе(минимальная стоимость не является числом)`)
			return
		}
		FilterStrings = append(FilterStrings, ft)
	}
	if len(keys[`price_max`]) > 0 {
		ft.Filter = "price <= "
		ft.Value, err = strconv.Atoi(keys[`price_max`][0])
		if err != nil {
			res = result.SetErrorResult(`Ошибка в запросе(максимальная стоимость не является числом)`)
			return
		}
		FilterStrings = append(FilterStrings, ft)
	}
	if len(keys[`cost_min`]) > 0 {
		ft.Filter = "cost >= "
		ft.Value, err = strconv.Atoi(keys[`cost_min`][0])
		if err != nil {
			res = result.SetErrorResult(`Ошибка в запросе(минимальная цена не является числом)`)
			return
		}
		FilterStrings = append(FilterStrings, ft)
	}
	if len(keys[`cost_max`]) > 0 {
		ft.Filter = "cost <= "
		ft.Value, err = strconv.Atoi(keys[`cost_max`][0])
		if err != nil {
			res = result.SetErrorResult(`Ошибка в запросе(максимальная цена не является числом)`)
			return
		}
		FilterStrings = append(FilterStrings, ft)
	}
	if len(keys[`cash_min`]) > 0 {
		ft.Filter = "cash >= "
		ft.Value, err = strconv.Atoi(keys[`cash_min`][0])
		if err != nil {
			res = result.SetErrorResult(`Ошибка в запросе(минимальные деньги в кассе не является числом)`)
			return
		}
		FilterStrings = append(FilterStrings, ft)
	}
	if len(keys[`cash_max`]) > 0 {
		ft.Filter = "cash <= "
		ft.Value, err = strconv.Atoi(keys[`cash_max`][0])
		if err != nil {
			res = result.SetErrorResult(`Ошибка в запросе(максимальные деньги в кассе не является числом)`)
			return
		}
		FilterStrings = append(FilterStrings, ft)
	}
	if len(keys[`str_min`]) > 0 {
		ft.Filter = "str >= "
		ft.Value, err = strconv.Atoi(keys[`str_min`][0])
		if err != nil {
			res = result.SetErrorResult(`Ошибка в запросе(минимальная сила не является числом)`)
			return
		}
		FilterStrings = append(FilterStrings, ft)
	}
	if len(keys[`str_max`]) > 0 {
		ft.Filter = "str <= "
		ft.Value, err = strconv.Atoi(keys[`str_max`][0])
		if err != nil {
			res = result.SetErrorResult(`Ошибка в запросе(максимальная сила не является числом)`)
			return
		}
		FilterStrings = append(FilterStrings, ft)
	}
	if len(keys[`country`]) > 0 {
		ft.Filter = "country = "
		ft.StringValues = s.Split(keys[`country`][0], ",")
		for i := 0; i < len(ft.StringValues); i++ {
			query := `SELECT name FROM list.nation_list where id = $1`
			ValueInt, err := strconv.Atoi(ft.StringValues[i])
			if err != nil {
				report.ErrorServer(r, err)
				res = result.SetErrorResult(`Ошибка в запроса(страны должны быть числами)`)
				return
			}
			params := []interface{}{ValueInt}
			err = db.QueryRowContext(ctx, query, params...).Scan(&ft.StringValues[i])
			if err != nil {
				report.ErrorSQLServer(r, err, query, params...)
				res = result.SetErrorResult(report.UnknownError)
				return
			}
			if ft.StringValues[i] == "" {
				res = result.SetErrorResult(`Ошибка в запросе(данной страны не существует)`)
				return
			}
		}
		FilterStrings = append(FilterStrings, ft)
	}
	res.Done = true
	return
}
