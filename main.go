package main

import (
	"flag"
	"io/ioutil"
	"gopkg.in/v1/yaml"
	"fmt"
	"mgun/work"
	"mgun/gun"
)

const (
	EMPTY_SIGN = ""
)

func main() {
	var file string
	flag.StringVar(&file, "f", EMPTY_SIGN, "/path/to/config/file.yaml")
	flag.Parse()

	if len(file) > 0 {
		bytes, err := ioutil.ReadFile(file)
		if err == nil {
			target := new(work.Target)
			err := yaml.Unmarshal(bytes, target)
			if err == nil {
				gun.Shoot(target)
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}
}

