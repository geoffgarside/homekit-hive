package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/geoffgarside/homekit-hive/pkg/api/v6/hive"
)

func main() {
	var (
		username string
		password string
		setTemp  float64
	)

	flag.StringVar(&username, "username", "", "hive username")
	flag.StringVar(&password, "password", "", "hive password")
	flag.StringVar(&username, "u", "", "hive username")
	flag.StringVar(&password, "p", "", "hive password")
	flag.Float64Var(&setTemp, "set", 0, "Set temperature")
	flag.Parse()

	c, err := hive.Connect(hive.WithCredentials(username, password))
	if err != nil {
		log.Fatal(err)
	}

	ts, err := c.Thermostats()
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range ts {
		currentTemp, err := t.Temperature()
		if err != nil {
			log.Print(err)
		}

		targetTemp, err := t.Target()
		if err != nil {
			log.Print(err)
		}

		fmt.Printf("%v\t%v\t%v\t%v\n",
			t.ID, t.Name, currentTemp, targetTemp)
	}

	if setTemp > 0 {
		if err := ts[0].SetTarget(setTemp); err != nil {
			log.Fatal(err)
		}
	}
}
