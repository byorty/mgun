package work

import (
	"fmt"
	"time"
)

const (
	HTTP_SCHEME = "http"
	GET_METHOD = "GET"
	POST_METHOD = "POST"
	PUT_METHOD = "PUT"
	DELETE_METHOD = "DELETE"
)

type Target struct {
	preparedHost string
	Concurrency  int
	LoopCount    int                     `yaml:"loopCount"`
	Timeout      time.Duration
	Scheme       string
	Host         string
	Port         int
	Headers      map[string] interface{}
	Shots       []*Shot                `yaml:"requests"`
}

func (this *Target) GetScheme() string {
	if len(this.Scheme) == 0 {
		this.Scheme = HTTP_SCHEME
	}
	return this.Scheme
}

func (this *Target) GetHost() string {
	if len(this.preparedHost) == 0 {
		if this.GetPort() != 80 {
			this.preparedHost = fmt.Sprintf("%s:%d", this.Host, this.Port)
		} else {
			this.preparedHost = this.Host
		}
	}
	return this.preparedHost
}

func (this *Target) GetPort() int {
	if this.Port == 0 {
		this.Port = 80
	}
	return this.Port
}

func (this *Target) GetConcurrency() int {
	if this.Concurrency == 0 {
		this.Concurrency = 1
	}
	return this.Concurrency
}

func (this *Target) GetLoopCount() int {
	if this.LoopCount == 0 {
		this.LoopCount = 1
	}
	return this.LoopCount
}

func (this *Target) GetTimeout() time.Duration {
	if this.Timeout == 0 {
		this.Timeout = 2
	}
	return this.Timeout
}
