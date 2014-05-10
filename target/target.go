package target

import (
	"fmt"
	"time"
	"errors"
)

const (
	HTTP_SCHEME = "http"
	HTTPS_SCHEME = "https"
	GET_METHOD = "GET"
	POST_METHOD = "POST"
	PUT_METHOD = "PUT"
	DELETE_METHOD = "DELETE"
)

func New() *Target {
	return new(Target)
}

type Target struct {
	Concurrency  int
	LoopCount    int                     `yaml:"loopCount"`
	Timeout      time.Duration
	Scheme       string
	Host         string
	Port         int
	Headers      map[string] interface{}
	Shots       []*Shot                  `yaml:"requests"`
}

func (this *Target) Check() error {
	if len(this.Scheme) > 0 && (this.Scheme != HTTP_SCHEME && this.Scheme != HTTPS_SCHEME) {
		return errors.New("invalid scheme")
	}

	if len(this.Host) == 0 {
		return errors.New("invalid host")
	}

	if len(this.Scheme) == 0 {
		this.Scheme = HTTP_SCHEME
	}

	if this.Port == 0 {
		this.Port = 80
	}

	if this.Port != 80 {
		this.Host = fmt.Sprintf("%s:%d", this.Host, this.Port)
	}

	if this.LoopCount == 0 {
		this.LoopCount = 1
	}

	if this.Concurrency == 0 {
		this.Concurrency = 1
	}

	if this.Timeout == 0 {
		this.Timeout = 2
	}

	if len(this.Shots) == 0 {
		this.Shots = append(this.Shots, &Shot{Get: "/"})
	}

	return nil
}
