package report

import (
	"fmt"
	"hm2/config"
	"net/http"
	"net/http/httputil"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kaibox-git/sqlparams"
	"github.com/lib/pq"
)

var UnknownError string = `Неизвестная ошибка`
var CtxError string = `Ошибка контекста`

func UserLog(m string, User config.User) {
	config.UserLogs[User.ID].Printf(m)
}

func ErrorServer(r *http.Request, err error) {
	var FileWithLineNum = FileWithLineNum()
	fmt.Printf("\n%s\n%s\n", FileWithLineNum, err.Error())
	m := createMessage(r, FileWithLineNum, err, "")
	go logError(m)
}

func ErrorSQLServer(r *http.Request, err error, query string, params ...interface{}) {
	var FileWithLineNum = FileWithLineNum()
	if err == nil {
		fmt.Printf("\n%s\n%s\n", FileWithLineNum, sqlparams.Inline(query, params...))
		return
	}
	fmt.Printf("\n%s:\n%s\n%s\n", FileWithLineNum, err.Error(), sqlparams.Inline(query, params...))
	m := createMessage(r, FileWithLineNum, err, sqlparams.Inline(query, params...))
	go logError(m)
}

func FatalError(r *http.Request, err error) {
	var FileWithLineNum = FileWithLineNum()
	fmt.Printf("\n%s\n%s\n", FileWithLineNum, err.Error())
	m := createMessage(r, FileWithLineNum, err, "")
	go logFatal(m)
}

func logFatal(m string) {
	config.ErrorLog.Print("FATAL ")
	config.ErrorLog.Printf("%s", m)
	config.ErrorLog.Println(strings.Repeat("—", 70))
	config.ErrorLog.Fatal("")
}

func logError(m string) {
	config.ErrorLog.Printf("%s", m)
	config.ErrorLog.Println(strings.Repeat("—", 70))
	config.ErrorLog.Println("")
}

func FileWithLineNum() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}
	return ``
}

func createMessage(r *http.Request, FileWithLineNum string, err error, sql string) string {
	var sb strings.Builder

	if len(sql) > 0 {
		var pqError *pq.Error
		if pq_err, ok := err.(*pq.Error); ok {
			pqError = pq_err
		} else {
			pqError = &pq.Error{}
		}
		fmt.Fprintf(&sb, "%s\n%s\n%s\n\n%s\n\n%#v\n\n%s",
			time.Now().Format("02.01.2006 15:04:05"),
			FileWithLineNum,
			err.Error(),
			sql,
			pqError,
			requestData(r))
	} else {
		fmt.Fprintf(&sb, "%s\n%s\n%s\n\n%s",
			time.Now().Format("02.01.2006 15:04:05"),
			FileWithLineNum,
			err.Error(),
			requestData(r))
	}
	return sb.String()
}

func requestData(r *http.Request) string {
	if r == nil {
		return ``
	}
	var (
		bBody       bool
		requestDump []byte
		dumpErr     error
	)
	// Чтобы httputil.DumpRequest работал с true (давал инфо о Body запроса), нужно в обработчике восстанавливать r.Body поcле
	// прочтения:
	// Например, в обработчике маршурута (для application/json):
	// b, _ := io.ReadAll(r.Body)
	// Тут же восстановление содержания r.Body, иначе оно пустое после прочтения выше.
	// r.Body = io.NopCloser(bytes.NewReader(b))

	if r.Header.Get("Content-Type") == `application/json` {
		bBody = true
	}
	requestDump, dumpErr = httputil.DumpRequest(r, bBody)
	if dumpErr != nil {
		requestDump = []byte(fmt.Sprintf("%s\ndumpError: %s\n", string(requestDump), dumpErr.Error()))
	}
	var sb strings.Builder
	sb.WriteString(strings.TrimSpace(string(requestDump)))
	sb.WriteString(postData(r))
	if sb.Len() > 0 {
		sb.WriteString("\n\n")
	}
	return sb.String()
}

func postData(r *http.Request) string {
	sb := strings.Builder{}
	if r != nil && r.Method == `POST` {
		r.ParseForm()
		if len(r.Form) > 0 {
			fmt.Fprintf(&sb, "\n\n========== POST DATA ========\n%#v\n\n", r.Form.Encode())
			// Сортировка по ключу по алфавиту
			var keys []string
			for k := range r.Form {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			// Итерация по отсортированному ключу:
			for _, key := range keys {
				for _, value := range r.Form[key] {
					fmt.Fprintf(&sb, "POST[%s]=%v\n", key, value)
				}
			}
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
