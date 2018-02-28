package main

import (
	"flag"
	"io/ioutil"
	"fmt"
	"./lib"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	var file string
	flag.StringVar(&file, "f", "/path/to/config/file.yaml", "configuration yaml file")
	flag.Parse()

	if len(file) > 0 {
		bytes, err := ioutil.ReadFile(file)
		if err == nil {

			kill := lib.GetKill()
			victim := lib.NewVictim()
			gun := lib.GetGun()
			reporter := lib.GetReporter()
			err := yaml.Unmarshal(bytes, kill)
			if err == nil {
				err = yaml.Unmarshal(bytes, victim)
				if err == nil {
					err = yaml.Unmarshal(bytes, reporter)
					if err == nil {
						err = yaml.Unmarshal(bytes, gun)
						if err == nil {
							kill.SetVictim(victim)
							kill.SetGun(gun)
							kill.Prepare()
							kill.Start()
						} else {
							fmt.Println(err)
						}
					} else {
						fmt.Println(err)
					}
				} else {
					fmt.Println(err)
				}
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println("config file not found")
	}
}
