package daemon

import (
	"bufio"
	"fmt"
	"hm2/convert"
	"hm2/teams"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func CommandLineStart() {
	var s string
	reader := bufio.NewReader(os.Stdin)
	for {
		s, _ = reader.ReadString('\n')
		if runtime.GOOS == "windows" {
			s = strings.Replace(s, "\r\n", "", -1)
		} else {
			s = strings.Replace(s, "\n", "", -1)
		}
		if s != `` {
			go CommandLineWorker(s)
		}
	}
}

func CommandLineWorker(s string) {
	Task := strings.Split(s, " ")
	switch len(Task) {
	case 0:
		IncorrectUsage()
	case 1:
		if strings.ToLower(Task[0]) == `help` {
			Help()
		} else {
			IncorrectUsage()
		}
	case 3:
	case 4:
		if Task[0] == `create` && Task[1] == `teams` {
			NumOfTeams, err := strconv.Atoi(Task[2])
			if err != nil {
				IncorrectUsage()
			}
			var data teams.CreateTeamData
			data.Country, err = convert.NationShortStringToLongString(Task[3])
			if err != nil {
				IncorrectCountry()
			}
			for i := 0; i < NumOfTeams; i++ {
				data.City = "Test" + strconv.Itoa(i)
				data.Stadium = "Test" + strconv.Itoa(i)
				data.Name = "TestName" + strconv.Itoa(i)
				teams.CreateTeamConfirm(nil, data, true)
			}
			Success()
		}
	default:
		IncorrectUsage()
	}

}

func Help() {
	fmt.Printf("List of all commands: \ncreate teans 10(num) RUS(country)")
}

func Success() {
	fmt.Printf("Completed!\n")
}

func IncorrectUsage() {
	fmt.Printf("Incorrect usage, please try 'help'\n")
}

func IncorrectCountry() {
	fmt.Printf("Incorrect country \n")
}
