package check

import (
	"strconv"
)

func Paginator(keys map[string][]string) (mes string, page int, limit int) {
	var err error
	if len(keys[`page`]) > 0 {
		page, err = strconv.Atoi(keys[`page`][0])
		if err != nil {
			mes = `Номер страницы должен быть числом`
			return mes, -1, -1
		}
	} else {
		mes = `Необходимо задать номер страницы`
		return mes, -1, -1
	}
	if len(keys[`limit`]) > 0 {
		limit, err = strconv.Atoi(keys[`limit`][0])
		if err != nil {
			mes = `Лимит должен быть числом`
			return mes, -1, -1
		}
	} else {
		mes = `Необходимо задать число элементов`
		return mes, -1, -1
	}
	return ``, page, limit
}
