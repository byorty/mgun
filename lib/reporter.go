package lib

import (
	"fmt"
	"time"
	hm "github.com/dustin/go-humanize"
	tm "github.com/buger/goterm"
	"math"
	"github.com/cznic/mathutil"
)

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

func (this *Reporter) report(kill *Kill, hits <- chan *Hit) {
	var startTime int64
	var endTime   int64
	var lastTime  time.Time
	var i         int
	requestsPerSeconds := make(map[int]map[int]int)
	reports := make(map[int]*ShotReport)
	hitsTable := tm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(hitsTable, "#\tRequest\tCompl.\tFail.\tMin.\tMax.\tAvg.\tAvail.\tMin, avg, max rps.\tContent len.\tTotal trans.\n")
	for hit := range hits {
		if startTime == 0 {
			startTime = hit.startTime.Unix()
		} else {
			startTime = mathutil.MinInt64(startTime, hit.startTime.Unix())
		}
		key := hit.shot.cartridge.id
		if report, ok := reports[key]; ok {
			report.update(hit)
		} else {
			report := NewShotReport(hit)
			reports[key] = report
		}

		if hit.startTime.Sub(lastTime) >= time.Second {
			lastTime = hit.endTime
			i++
		}

		if _, ok := requestsPerSeconds[i]; ok {
			requestsPerSeconds[i][hit.shot.cartridge.id]++
		} else {
			requestsPerSeconds[i] = make(map[int]int)
			requestsPerSeconds[i][hit.shot.cartridge.id] = 1
		}

		if endTime < 0 {
			endTime = hit.endTime.Unix()
		} else {
			endTime = mathutil.MaxInt64(endTime, hit.endTime.Unix())
		}
	}

	var totalRequests          int
	var completeRequests       int
	var failedRequests         int
	var availability           float64
	var totalRequestPerSeconds float64
	var totalTransferred       int64

	reportsCount := float64(len(reports))
	cartridges := kill.gun.Cartridges.toPlainSlice()
	for _, cartridge := range cartridges {

		if report, ok := reports[cartridge.id]; ok {
			counts := make([]int, 0)
			for _, countById := range requestsPerSeconds {
				if count, ok := countById[cartridge.id]; ok {
					counts = append(counts, count)
				}
			}
			var minRequestPerSecond int64
			var avgRequestPerSecond float64
			var maxRequestPerSecond int64
			for _, count := range counts {
				count64 := int64(count)
				if minRequestPerSecond == 0 {
					minRequestPerSecond = count64
				} else {
					minRequestPerSecond = mathutil.MinInt64(minRequestPerSecond, count64)
				}
				avgRequestPerSecond += float64(count)
				if maxRequestPerSecond == 0 {
					maxRequestPerSecond = count64
				} else {
					maxRequestPerSecond = mathutil.MaxInt64(maxRequestPerSecond, count64)
				}
			}
			avgRequestPerSecond = avgRequestPerSecond / float64(len(counts))

			name := this.getRequestName(cartridge)
			totalRequests += report.totalRequests
			completeRequests += report.completeRequests
			failedRequests += report.failedRequests
			availability += report.getAvailability()
			totalTransferred += report.totalTransferred
			totalRequestPerSeconds += avgRequestPerSecond

			fmt.Fprintf(
				hitsTable, "%d.\t%s\t%d\t%d\t%.3fs.\t%.3fs.\t%.3fs.\t%.2f%%\t%d / ~ %.2f / %d\t%s\t%s\n",
				cartridge.id,
				name,
				report.completeRequests,
				report.failedRequests,
				report.minTime,
				report.maxTime,
				report.getAvgTime(),
				report.getAvailability(),
				minRequestPerSecond,
				avgRequestPerSecond,
				maxRequestPerSecond,
				hm.Bytes(uint64(report.contentLength)),
				hm.Bytes(uint64(report.totalTransferred)),
			)
		}
	}

	targetTable := tm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(targetTable, "Server Hostname:\t%s\n", kill.victim.Host)
	fmt.Fprintf(targetTable, "Server Port:\t%d\n", kill.victim.Port)
	fmt.Fprintf(targetTable, "Concurrency Level:\t%d\n", kill.GunsCount)
	fmt.Fprintf(targetTable, "Loop count:\t%d\n", kill.AttemptsCount)
	fmt.Fprintf(targetTable, "Timeout:\t%d seconds\n", kill.Timeout)
	fmt.Fprintf(targetTable, "Time taken for tests:\t%d seconds\n", int(time.Unix(endTime, 0).Sub(time.Unix(startTime, 0)).Seconds()))
	fmt.Fprintf(targetTable, "Total requests:\t%d\n", totalRequests)
	fmt.Fprintf(targetTable, "Complete requests:\t%d\n", completeRequests)
	fmt.Fprintf(targetTable, "Failed requests:\t%d\n", failedRequests)
	fmt.Fprintf(targetTable, "Availability:\t%.2f%%\n", availability / reportsCount)
	fmt.Fprintf(targetTable, "Requests per second:\t~ %.2f\n", totalRequestPerSeconds / float64(len(cartridges)))
	fmt.Fprintf(targetTable, "Total transferred:\t%s\n", hm.Bytes(uint64(totalTransferred)))

	fmt.Println(EMPTY_SIGN)
	fmt.Println(EMPTY_SIGN)
	fmt.Println(targetTable)
	fmt.Println(hitsTable)
}

func (this *Reporter) getRequestName(cartridge *Cartridge) string {
	return fmt.Sprintf("%s %s", cartridge.getMethod(), cartridge.path.rawDescription)
}

func NewShotReport(hit *Hit) *ShotReport {
	return new(ShotReport).create(hit)
}

type ShotReport struct {
	totalRequests     int
	startTime         time.Time
	endTime           time.Time
	minTime           float64
	maxTime           float64
	completeRequests  int
	failedRequests    int
	requestsPerSecond float64
	totalTransferred  int64
	totalTime         float64
	contentLength     int64
}

func (this *ShotReport) create(hit *Hit) *ShotReport {
	timeRequest := this.getDiffSeconds(hit)
	this.minTime = timeRequest
	this.maxTime = timeRequest
	this.totalTime = timeRequest
	this.updateTotalRequests()
	this.updateTotalTransferred(hit)
	this.checkResponseStatusCode(hit)
	this.startTime = hit.startTime
	this.endTime = hit.endTime
	return this
}

func (this *ShotReport) getDiffSeconds(hit *Hit) float64 {
	return hit.endTime.Sub(hit.startTime).Seconds()
}

func (this *ShotReport) checkResponseStatusCode(hit *Hit) {
	shot := hit.shot
	if hit.shot.request != nil && hit.response != nil {
		statusCode := hit.response.StatusCode
		if this.inArray(statusCode, shot.cartridge.failedStatusCodes) {
			this.failedRequests++
		} else if this.inArray(statusCode, shot.cartridge.successStatusCodes) {
			this.completeRequests++
		} else {
			this.failedRequests++
		}
	} else {
		this.failedRequests++
	}
}

func (this *ShotReport) inArray(a int, array []int) bool {
	for _, b := range array {
		if a == b {
			return true
		}
	}
	return false
}

func (this *ShotReport) updateTotalRequests() {
	this.totalRequests++
}

func (this *ShotReport) updateTotalTransferred(hit *Hit) {
	if hit.response != nil {
		this.totalTransferred += int64(len(hit.responseBody))
		if this.contentLength == 0 {
			this.contentLength = this.totalTransferred
		}
	}
}

func (this *ShotReport) updateRequestsPerSecond(timeRequest float64) {
	if timeRequest == 0 {
		timeRequest = 1
	}
	if this.requestsPerSecond == 0 {
		this.requestsPerSecond = 1 / timeRequest
	} else {
		this.requestsPerSecond = ((1 / timeRequest) + this.requestsPerSecond) / 2
	}
	reporter.log("time request: %v, requests per second: %v, avg requests per second: %v", timeRequest, 1 / timeRequest, this.requestsPerSecond)
}

func (this *ShotReport) update(hit *Hit) *ShotReport {
	timeRequest := this.getDiffSeconds(hit)
	this.minTime = math.Min(this.minTime, timeRequest)
	this.maxTime = math.Max(this.maxTime, timeRequest)
	this.totalTime += timeRequest
	this.updateTotalRequests()
	this.updateTotalTransferred(hit)
	this.checkResponseStatusCode(hit)
	this.endTime = hit.endTime
	return this
}

func (this *ShotReport) getAvgTime() float64 {
	return (this.minTime + this.maxTime) / 2
}

func (this *ShotReport) getAvailability() float64 {
	return float64(this.completeRequests) * 100 / float64(this.totalRequests)
}
