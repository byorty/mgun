package main

import (
	"flag"
	"io/ioutil"
	"gopkg.in/v1/yaml"
	"fmt"
	"github.com/byorty/mgun/target"
	"github.com/byorty/mgun/gun"
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
			newTarget := target.New()
			err := yaml.Unmarshal(bytes, newTarget)
			if err == nil {
				err := newTarget.Check()
				if err == nil {
					gun.Shoot(newTarget)
				}
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}
}

