package daemon

import "fmt"

func CommandLineStart() {
	var s string
	for {
		fmt.Scanln(&s)
		go CommandLineWorker(s)
	}
}

func CommandLineWorker(s string) {
	fmt.Println(s)
}
