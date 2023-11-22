package main

import (
	"flag"
	"log"
	"strings"
)

var (
	serverList string
	port       int
)

func init() {
	flag.StringVar(
		&serverList,
		"backends",
		"http://localhost:3031,http://localhost:3032,http://localhost:3033,http://localhost:3034",
		"Load balanced backends, use commas to separate",
	)
	flag.IntVar(&port, "port", 3030, "Port to serve")
}

func main() {
	flag.Parse()
	if len(serverList) == 0 {
		log.Fatal("предоставьте один или несколько бэкэндов для балансировки нагрузки")
	}

	pool, err := NewPool(strings.Split(serverList, ","))
	if err != nil {
		log.Fatal(err)
	}

	err = pool.Start(port)
	if err != nil {
		log.Fatal(err)
	}
}
