package gun

import (
	"mgun/work"
	"net/http"
	"net/url"
	"reflect"
	"bytes"
	"mime/multipart"
	"strings"
	"net/http/cookiejar"
	"code.google.com/p/go.net/publicsuffix"
	"fmt"
	"net"
	"time"
)

type Cage struct {
	target *work.Target
	bullets chan <- *work.Bullet
}

func (this *Cage) SetTarget(target *work.Target) *Cage {
	this.target = target
	return this
}

func (this *Cage) SetBullets(bullets chan <- *work.Bullet) *Cage {
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
	client.Transport = &http.Transport{
		Dial: func(network, addr string) (conn net.Conn, err error) {
			return net.DialTimeout(network, addr, time.Second * this.target.GetTimeout())
		},
		ResponseHeaderTimeout: time.Second * this.target.GetTimeout(),
	}
	client.Jar = jar
	for _, shot := range this.target.Shots {

		bullet := new(work.Bullet)
		bullet.Shot = shot
		bullet.Client = client

		reqUrl := new(url.URL)
		reqUrl.Scheme = this.target.GetScheme()
		reqUrl.Host = this.target.GetHost()

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
