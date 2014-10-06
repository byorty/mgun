package mgun

import "fmt"

const (
	EMPTY_SIGN = ""
)

var (
	reporter = new(Reporter)
)

func GetReporter() *Reporter {
	return reporter
}

type Reporter struct {
	Debug bool `yaml:"debug"`
}

func (this *Reporter) log(message string, args ...interface{}) {
	if this.Debug {
		message = fmt.Sprintf(message, args...)
		fmt.Println(message)
	}
}

func (this *Reporter) ln() {
	this.log(EMPTY_SIGN)
}
