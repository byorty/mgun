package mgun

import (
	"time"
	"errors"
	"fmt"
	"runtime"
	"net/http"
	"code.google.com/p/go.net/publicsuffix"
	"net/http/cookiejar"
	"net"
	"strings"
	"net/url"
)

var (
	kill = &Kill{shotsCount: 0}
)

type Kill struct {
	shotsCount    int
	GunsCount     int           `yaml:"concurrency"`
	AttemptsCount int           `yaml:"loopCount"`
	Timeout       time.Duration `yaml:"timeout"`
	gun           *Gun
	victim        *Victim
}

func GetKill() *Kill {
	return kill
}

func (this *Kill) SetGun(gun *Gun) {
	this.gun = gun
}

func (this *Kill) SetVictim(victim *Victim) {
	this.victim = victim
}

func (this *Kill) Prepare() error {
	reporter.ln()
	reporter.log("prepare kill")

	err := this.victim.prepare()
	this.gun.prepare()

	if this.GunsCount == 0 {
		this.GunsCount = 1
	}
	reporter.log("guns count - %v", this.GunsCount)

	if this.AttemptsCount == 0 {
		this.AttemptsCount = 1
	}
	reporter.log("attempts count - %v", this.AttemptsCount)

	if this.Timeout == 0 {
		this.Timeout = 2
	}
	reporter.log("timeout - %v", this.GunsCount)
	reporter.log("shots count - %v", this.shotsCount)

	return err
}

func (this *Kill) Start() {
	reporter.ln()
	reporter.log("start kill")

	// отдаем рутинам все ядра процессора
	runtime.GOMAXPROCS(runtime.NumCPU())
	// считаем кол-во результатов
//	hitCount := this.GunsCount * this.AttemptsCount * this.shotsCount

	// создаем програсс бар
//	bar := pb.StartNew(hitCount)
//	group := new(sync.WaitGroup)
	// создаем канал результатов
//	hits := make(chan *Hit, hitCount)

	// запускаем повторения заданий, если в настройках не указано кол-во повторений,
	// тогда программа сделает одно повторение
	for i := 0; i < this.AttemptsCount; i++ {
		reporter.log("attempt - %v", i)
//		group.Add(hitCount / this.AttemptsCount)
		// запускаем конкуретные задания, если в настройках не указано кол-во заданий,
		// тогда программа сделает одно задание
		for j := 0; j < this.GunsCount; j++ {

//			shots := make(chan *Shot, this.shotsCount)

//			bullets := make(chan *target.Bullet, shotCount)

			killer := new(Killer)
			killer.SetVictim(this.victim)
			killer.SetGun(this.gun)

//			worker := new(Gun).
//			SetGroup(group).
//			SetProgressBar(bar).
//			SetHits(hits).
//			SetBullets(bullets).
//			SetTarget(newTarget)
//			go worker.Fire()
//			// создаем запросы
//			cage := new(Cage).
//			SetBullets(bullets).
//			SetTarget(newTarget)
//			go killer.Charge()
			reporter.log("killer - %v charge", j)
			killer.Charge()
		}

//		group.Wait()
	}

//	close(hits)
	// аггрегируем результаты задания и выводим статистику в консоль
//	reporter.report(newTarget, hits)
}

type Shot struct {
	cartridge *Cartridge
	request *http.Request
	client *http.Client
	transport *http.Transport
}

type Killer struct {
	victim *Victim
	gun    *Gun
}

func (this *Killer) SetVictim(victim *Victim) {
	this.victim = victim
}

func (this *Killer) SetGun(gun *Gun) {
	this.gun = gun
}

func (this *Killer) Charge() {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		reporter.log("cookie don't created - %v", err)
	}
	client := new(http.Client)
	client.Jar = jar
	for _, cartridge := range this.gun.Cartridges {

		var timeout time.Duration
		if cartridge.timeout > 0 {
			timeout = cartridge.timeout
		} else {
			timeout = kill.Timeout
		}

		shot := new(Shot)
		shot.cartridge = cartridge
		shot.client = client
		shot.transport = &http.Transport{
			Dial: func(network, addr string) (conn net.Conn, err error) {
				return net.DialTimeout(network, addr, time.Second * timeout)
			},
			ResponseHeaderTimeout: time.Second * timeout,
		}

		reqUrl := new(url.URL)
		reqUrl.Scheme = this.victim.Scheme
		reqUrl.Host = this.victim.Host

		path := cartridge.GetPath().String()
		pathParts := strings.Split(path, "?")
		reqUrl.Path = pathParts[0]
		if len(pathParts) == 2 {
			val, _ := url.ParseQuery(pathParts[1])
			reqUrl.RawQuery = val.Encode()
		} else {
			reqUrl.RawQuery = ""
		}

//
//		var body bytes.Buffer
//
//		var writer *multipart.Writer
//		if shot.IsPost() {
//			writer = multipart.NewWriter(&body)
//			for key, value := range shot.Params {
//				writer.WriteField(key, reflect.ValueOf(value).String())
//			}
//			writer.Close()
//		}

		reporter.log("create request - %v", reqUrl)

	}
}

type Hit struct {
	StartTime time.Time
	EndTime time.Time
//	Shot *Shot
	Request *http.Request
	Response *http.Response
	ResponseBody []byte
}

const (
	HTTP_SCHEME = "http"
	HTTPS_SCHEME = "https"
)

type Victim struct {
	Scheme string `yaml:"scheme"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
}

func NewVictim() *Victim {
	return new(Victim)
}

func (this *Victim) prepare() error {
	if len(this.Scheme) > 0 && (this.Scheme != HTTP_SCHEME && this.Scheme != HTTPS_SCHEME) {
		return errors.New("invalid scheme")
	}

	if len(this.Host) == 0 {
		return errors.New("invalid host")
	}

	if len(this.Scheme) == 0 {
		this.Scheme = HTTP_SCHEME
	}
	reporter.log("scheme - %v", this.Scheme)

	if this.Port == 0 {
		this.Port = 80
	}
	reporter.log("port - %v", this.Port)

	if this.Port != 80 {
		this.Host = fmt.Sprintf("%s:%d", this.Host, this.Port)
	}
	reporter.log("host - %v", this.Host)

	return nil
}
