package gun

import (
	"sync"
	"time"
	"mgun/target"
	"github.com/cheggaaa/pb"
	"runtime"
)

func Shoot(newTarget *target.Target) {
	// создаем докладчика
	reporter := NewReporter()

	// отдаем рутинам все ядра процессора
	runtime.GOMAXPROCS(runtime.NumCPU())

	shotCount := len(newTarget.Shots)
	// считаем кол-во результатов
	hitCount := newTarget.Concurrency * newTarget.LoopCount * shotCount

	// создаем програсс бар
	bar := pb.StartNew(hitCount)
	group := new(sync.WaitGroup)
	// создаем канал результатов
	hits := make(chan *target.Hit, hitCount)

	// запускаем повторения заданий, если в настройках не указано кол-во повторений,
	// тогда программа сделает одно повторение
	for i := 0; i < newTarget.LoopCount; i++ {
		group.Add(hitCount / newTarget.LoopCount)
		// запускаем конкуретные задания, если в настройках не указано кол-во заданий,
		// тогда программа сделает одно задание
		for j := 0; j < newTarget.Concurrency; j++ {
			bullets := make(chan *target.Bullet, shotCount)

			worker := new(Gun).
				SetGroup(group).
				SetProgressBar(bar).
				SetHits(hits).
				SetBullets(bullets)
			go worker.Fire()
			// создаем запросы
			cage := new(Cage).
				SetBullets(bullets).
				SetTarget(newTarget)
			go cage.Charge()
		}

		group.Wait()
	}

	close(hits)
	// аггрегируем результаты задания и выводим статистику в консоль
	reporter.report(newTarget, hits)
}

type Gun struct {
	bullets <- chan *target.Bullet
	hits chan <- *target.Hit
	group *sync.WaitGroup
	bar *pb.ProgressBar
}

func (this *Gun) SetBullets(bullets <- chan *target.Bullet) *Gun {
	this.bullets = bullets
	return this
}

func (this *Gun) SetHits(hits chan <- *target.Hit) *Gun {
	this.hits = hits
	return this
}
func (this *Gun) SetGroup(group *sync.WaitGroup) *Gun {
	this.group = group
	return this
}

func (this *Gun) SetProgressBar(bar *pb.ProgressBar) *Gun {
	this.bar = bar
	return this
}

func (this *Gun) Fire() {
	for bullet := range this.bullets {
		this.bar.Increment()
		hit := new(target.Hit)
		hit.Shot = bullet.Shot
		hit.Request = bullet.Request
		hit.StartTime = time.Now()
//		dump, _ := httputil.DumpRequest(bullet.Request, true)
//		fmt.Println(string(dump))
		bullet.Client.Transport = bullet.Transport
		resp, err := bullet.Client.Do(bullet.Request)
		if err == nil {
//			dump, _ := httputil.DumpResponse(resp, true)
//			fmt.Println(string(dump))
			resp.Body.Close()
			hit.Response = resp
		} else {
//			fmt.Println(err)
		}
		hit.EndTime = time.Now()
		this.hits <- hit
		this.group.Done()
	}
}
