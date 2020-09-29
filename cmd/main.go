package main

import (
	"fmt"
	"os"

	"github.com/scottyw/relic/relic"
)

func main() {
	if err := relic.Pick(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
