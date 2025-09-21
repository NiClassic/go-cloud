package config

import (
	"flag"
	"os"
)

var Debug bool

func Init() {
	flag.BoolVar(&Debug, "debug", false, "enable debug mode")
	flag.Parse()
	if !Debug {
		val, ok := os.LookupEnv("DEBUG")
		if ok {
			Debug = val == "true"
		}
	}
}
