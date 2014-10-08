package mgun

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
	var endTime int64
	requestsPerSeconds := make(map[int64]map[int]int)
	reports := make(map[int]*ShotReport)
	hitsTable := tm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(hitsTable, "#\tRequest\tCompl.\tFail.\tMin.\tMax.\tAvg.\tAvail.\tReq. per sec.\tContent len.\tTotal trans.\n")
	for hit := range hits {
		if startTime < 0 {
			startTime = hit.startTime.Unix()
		} else {
//			startTime = math.Min(startTime, hit.startTime.Unix())
		}
		key := hit.shot.cartridge.id
		if report, ok := reports[key]; ok {
			report.Update(hit)
		} else {
			report := NewShotReport(hit)
			reports[key] = report
		}

		if _, ok := requestsPerSeconds[hit.endTime.Unix()]; ok {
			requestsPerSeconds[hit.endTime.Unix()][hit.shot.cartridge.id]++
		} else {
			requestsPerSeconds[hit.endTime.Unix()] = make(map[int]int)
			requestsPerSeconds[hit.endTime.Unix()][hit.shot.cartridge.id] = 1
		}

		if endTime < 0 {
			endTime = hit.endTime.Unix()
		} else {
//			endTime = math.Max(endTime, hit.endTime.Unix())
		}
	}

//	fmt.Println(requestsPerSeconds)

	reportsCount := float64(len(reports))
	var totalRequests int
	var completeRequests int
	var failedRequests int
	var availability float64
	var totalRequestPerSeconds float64
	var totalTransferred int64

	cartridges := kill.gun.Cartridges.toPlainSlice()
	for _, cartridge := range cartridges {

		if report, ok := reports[cartridge.id]; ok {
			counts := make([]int, 0)
			for _, countById := range requestsPerSeconds {
				if count, ok := countById[cartridge.id]; ok {
					counts = append(counts, count)
				}
			}
			var requestPerSecond float64
			for _, count := range counts {
				requestPerSecond += float64(count)
			}
			requestPerSecond = requestPerSecond / float64(len(counts))

			name := this.getRequestName(cartridge)
			totalRequests += report.TotalRequests
			completeRequests += report.CompleteRequests
			failedRequests += report.FailedRequests
			availability += report.GetAvailability()
			totalTransferred += report.TotalTransferred
			totalRequestPerSeconds += requestPerSecond

			fmt.Fprintf(
				hitsTable, "%d.\t%s\t%d\t%d\t%.3fs.\t%.3fs.\t%.3fs.\t%.2f%%\t~ %.2f\t%s\t%s\n",
				cartridge.id,
				name,
				report.CompleteRequests,
				report.FailedRequests,
				report.MinTime,
				report.MaxTime,
				report.GetAvgTime(),
				report.GetAvailability(),
				requestPerSecond,
				hm.Bytes(uint64(report.ContentLength)),
				hm.Bytes(uint64(report.TotalTransferred)),
			)
		}
	}

	targetTable := tm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(targetTable, "Server Hostname:\t%s\n", kill.victim.Host)
	fmt.Fprintf(targetTable, "Server Port:\t%d\n", kill.victim.Port)
	fmt.Fprintf(targetTable, "Concurrency Level:\t%d\n", kill.GunsCount)
	fmt.Fprintf(targetTable, "Loop count:\t%d\n", kill.AttemptsCount)
	fmt.Fprintf(targetTable, "Timeout:\t%d seconds\n", kill.Timeout)
	fmt.Fprintf(targetTable, "Time taken for tests:\t%.3f seconds\n", endTime.Sub(startTime).Seconds())
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
	return fmt.Sprintf("%s %s", cartridge.GetMethod(), cartridge.GetPathAsString())
}

func NewShotReport(hit *Hit) *ShotReport {
	return new(ShotReport).create(hit)
}

type ShotReport struct {
	TotalRequests     int
	startTime         time.Time
	endTime           time.Time
	MinTime           float64
	MaxTime           float64
	CompleteRequests  int
	FailedRequests    int
	requestsPerSecond float64
	TotalTransferred  int64
	TotalTime         float64
	ContentLength     int64
}

func (this *ShotReport) create(hit *Hit) *ShotReport {
	timeRequest := this.getDiffSeconds(hit)
	this.MinTime = timeRequest
	this.MaxTime = timeRequest
	this.TotalTime = timeRequest
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
			this.FailedRequests++
		} else if this.inArray(statusCode, shot.cartridge.successStatusCodes) {
			this.CompleteRequests++
		} else {
			this.FailedRequests++
		}
	} else {
		this.FailedRequests++
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
	this.TotalRequests++
}

func (this *ShotReport) updateTotalTransferred(hit *Hit) {
	if hit.response != nil {
		this.TotalTransferred += int64(len(hit.responseBody))
		if this.ContentLength == 0 {
			this.ContentLength = this.TotalTransferred
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

func (this *ShotReport) Update(hit *Hit) *ShotReport {
	timeRequest := this.getDiffSeconds(hit)
	this.MinTime = math.Min(this.MinTime, timeRequest)
	this.MaxTime = math.Max(this.MaxTime, timeRequest)
	this.TotalTime += timeRequest
	this.updateTotalRequests()
	this.updateTotalTransferred(hit)
	this.checkResponseStatusCode(hit)
	this.endTime = hit.endTime
	return this
}

func (this *ShotReport) GetAvgTime() float64 {
	return (this.MinTime + this.MaxTime) / 2
}

func (this *ShotReport) GetAvailability() float64 {
	return float64(this.CompleteRequests) * 100 / float64(this.TotalRequests)
}
