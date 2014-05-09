package target

import "time"

type Shot struct {
	path    string
	method  string
	Get     string                  `yaml:"GET"`
	Post    string                  `yaml:"POST"`
	Put     string                  `yaml:"PUT"`
	Delete  string                  `yaml:"DELETE"`
	Headers map[string] interface{}
	Params  map[string] interface{}
	SuccessStatusCodes []int        `yaml:"successStatusCodes"`
	FailedStatusCodes  []int        `yaml:"failedStatusCodes"`
	Timeout time.Duration
}

func (this *Shot) GetMethod() string {
	if len(this.method) == 0 {
		if len(this.Get) > 0 {
			this.method = GET_METHOD
		} else if len(this.Post) > 0 {
			this.method = POST_METHOD
		} else if len(this.Put) > 0 {
			this.method = PUT_METHOD
		} else if len(this.Delete) > 0 {
			this.method = DELETE_METHOD
		}
	}
	return this.method
}

func (this *Shot) GetPath() string {
	if len(this.path) == 0 {
		if len(this.Get) > 0 {
			this.path = this.Get
		} else if len(this.Post) > 0 {
			this.path = this.Post
		} else if len(this.Put) > 0 {
			this.path = this.Put
		} else if len(this.Delete) > 0 {
			this.path = this.Delete
		}
	}
	return this.path
}

func (this *Shot) IsPost() bool {
	return this.GetMethod() == POST_METHOD
}

func (this *Shot) GetSuccessStatusCodes() []int {
	if len(this.SuccessStatusCodes) == 0 {
		this.SuccessStatusCodes = []int{200}
	}
	return this.SuccessStatusCodes
}
