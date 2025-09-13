package config

import "flag"

var Debug bool

func Init() {
	flag.BoolVar(&Debug, "debug", false, "enable debug mode")
	flag.Parse()
}
