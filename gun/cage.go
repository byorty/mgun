package gun

import (
	"mgun/target"
	"net/http"
	"net/url"
	"reflect"
	"bytes"
	"mime/multipart"
	"strings"
	"net/http/cookiejar"
	"code.google.com/p/go.net/publicsuffix"
	"fmt"
	"time"
	"net"
)

type Cage struct {
	target *target.Target
	bullets chan <- *target.Bullet
}

func (this *Cage) SetTarget(target *target.Target) *Cage {
	this.target = target
	return this
}

func (this *Cage) SetBullets(bullets chan <- *target.Bullet) *Cage {
	this.bullets = bullets
	return this
}

func (this *Cage) Сharge() {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		fmt.Println(err)
	}
	client := new(http.Client)
	client.Jar = jar
	for _, shot := range this.target.Shots {

		var timeout time.Duration
		if shot.Timeout > 0 {
			timeout = shot.Timeout
		} else {
			timeout = this.target.Timeout
		}

		client.Transport = &http.Transport{
			Dial: func(network, addr string) (conn net.Conn, err error) {
				return net.DialTimeout(network, addr, time.Second * timeout)
			},
			ResponseHeaderTimeout: time.Second * timeout,
		}

		bullet := new(target.Bullet)
		bullet.Shot = shot
		bullet.Client = client

		reqUrl := new(url.URL)
		reqUrl.Scheme = this.target.Scheme
		reqUrl.Host = this.target.Host

		path := shot.GetPath()
		pathParts := strings.Split(path, "?")
		reqUrl.Path = pathParts[0]
		if len(pathParts) == 2 {
			val, _ := url.ParseQuery(pathParts[1])
			reqUrl.RawQuery = val.Encode()
		} else {
			reqUrl.RawQuery = ""
		}

		var body bytes.Buffer

		var writer *multipart.Writer
		if shot.IsPost() {
			writer = multipart.NewWriter(&body)
			for key, value := range shot.Params {
				writer.WriteField(key, reflect.ValueOf(value).String())
			}
			writer.Close()
		}

		request, err := http.NewRequest(shot.GetMethod(), reqUrl.String(), &body)

		this.setRequestHeaders(request, this.target.Headers)
		this.setRequestHeaders(request, shot.Headers)

		if shot.IsPost() {
			request.Header.Set("Content-Type", writer.FormDataContentType())
		}

		if err == nil {
			bullet.Request = request
		}
		this.bullets <- bullet
	}
	close(this.bullets)
}

func (this *Cage) setRequestHeaders(request *http.Request, headers map[string] interface{}) {
	for key, value := range headers {
		request.Header.Set(key, reflect.ValueOf(value).String())
	}
}
