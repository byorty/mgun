package main

import (
	"flag"
	"io/ioutil"
	"fmt"
	"github.com/byorty/mgun"
	yaml "gopkg.in/yaml.v2"
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

			victim := mgun.NewVictim()
			kill := mgun.GetKill()
			gun := mgun.GetGun()
			reporter := mgun.GetReporter()

			err := yaml.Unmarshal(bytes, kill)
			err = yaml.Unmarshal(bytes, victim)
			err = yaml.Unmarshal(bytes, reporter)
			err = yaml.Unmarshal(bytes, gun)

			kill.SetVictim(victim)
			kill.SetGun(gun)
			kill.Prepare()
			kill.Start()

//			fmt.Println(gun)
//			fmt.Println(gun.Cartridges)
			if err == nil {
				//				newTarget := target.New()
				//				if concurrency, ok := config["concurrency"]; ok {
				//					newTarget.SetConcurrency(concurrency.(int))
				//				}
				//
				//				if loopCount, ok := config["loopCount"]; ok {
				//					newTarget.SetLoopCount(loopCount.(int))
				//				}
				//
				//				if timeout, ok := config["timeout"]; ok {
				//					newTarget.SetTimeout(time.Duration(timeout.(int)))
				//				}
				//
				//				if scheme, ok := config["scheme"]; ok {
				//					newTarget.SetScheme(scheme.(string))
				//				}
				//
				//				if host, ok := config["host"]; ok {
				//					newTarget.SetHost(host.(string))
				//				}
				//
				//				if port, ok := config["port"]; ok {
				//					newTarget.SetPort(port.(int))
				//				}
				//
				//				if headers, ok := config["headers"]; ok {
				//					fillMap(headers.(map[interface{}] interface{}), newTarget.AddHeader)
				//				}

				//				if params, ok := config["params"]; ok {
				//					fmt.Println(params)
				//					paramsMap := params.(map[interface{}] interface{})
				//					session := paramsMap["session"]

				//					fmt.Println(session.([]map[interface{}] interface{}))
				//				}

				//				if requests, ok := config["requests"]; ok {
				//					newTarget.SetShots(createShots(requests.([]interface{})))
				//				} else {
				//					newTarget.SetShots([]*target.Shot{
				//						target.NewShot().
				//							SetMethod(target.GET_METHOD).
				//							SetPath("/"),
				//					})
				//				}
				//				fmt.Println(newTarget)
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}
}
