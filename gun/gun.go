package gun

import (
	"sync"
	"time"
	"mgun/work"
	"github.com/cheggaaa/pb"
	"runtime"
)

func Shoot(target *work.Target) {
	// создаем докладчика
	reporter := NewReporter()

	// отдаем рутинам все ядра процессора
	runtime.GOMAXPROCS(runtime.NumCPU())

	// считаем кол-во результатов
	hitCount := target.GetConcurrency() * target.GetLoopCount() * len(target.Shots)

	// создаем програсс бар
	bar := pb.StartNew(hitCount)
	group := new(sync.WaitGroup)
	// создаем канал результатов
	hits := make(chan *work.Hit, hitCount)

	// запускаем повторения заданий, если в настройках не указано кол-во повторений,
	// тогда программа сделает одно повторение
	for i := 0; i < target.GetLoopCount(); i++ {
		group.Add(hitCount / target.GetLoopCount())
		// запускаем конкуретные задания, если в настройках не указано кол-во заданий,
		// тогда программа сделает одно задание
		for j := 0; j < target.GetConcurrency(); j++ {
			bullets := make(chan *work.Bullet, len(target.Shots))

			worker := new(Gun).
				SetGroup(group).
				SetProgressBar(bar).
				SetHits(hits).
				SetBullets(bullets)
			go worker.Fire()
			// создаем запросы
			cage := new(Cage).
				SetBullets(bullets).
				SetTarget(target)
			go cage.Сharge()
		}

		group.Wait()
	}

	close(hits)
	// аггрегируем результаты задания и выводим статистику в консоль
	reporter.report(hits)
}

type Gun struct {
	bullets <- chan *work.Bullet
	hits chan <- *work.Hit
	group *sync.WaitGroup
	bar *pb.ProgressBar
}

func (this *Gun) SetBullets(bullets <- chan *work.Bullet) *Gun {
	this.bullets = bullets
	return this
}

func (this *Gun) SetHits(hits chan <- *work.Hit) *Gun {
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
		hit := new(work.Hit)
		hit.Shot = bullet.Shot
		hit.Request = bullet.Request
		hit.StartTime = time.Now()
//		dump, _ := httputil.DumpRequest(target.Request, true)
//		fmt.Println(string(dump))
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
