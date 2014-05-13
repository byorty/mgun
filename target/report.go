package target

import (
	"math"
	"time"
)

func NewStageReport(hit *Hit) *ShotReport {
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
	RequestsPerSecond float64
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
	this.startTime = hit.StartTime
	this.endTime = hit.EndTime
	return this
}

func (this *ShotReport) getDiffSeconds(hit *Hit) float64 {
	return hit.EndTime.Sub(hit.StartTime).Seconds()
}

func (this *ShotReport) checkResponseStatusCode(hit *Hit) {
	shot := hit.Shot
	if hit.Request != nil && hit.Response != nil {
		statusCode := hit.Response.StatusCode
		if this.inArray(statusCode, shot.FailedStatusCodes) {
			this.FailedRequests++
		} else if this.inArray(statusCode, shot.GetSuccessStatusCodes()) {
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
	if hit.Response != nil {
		this.TotalTransferred += int64(len(hit.ResponseBody))
		if this.ContentLength == 0 {
			this.ContentLength = this.TotalTransferred
		}
	}
}

func (this *ShotReport) Update(hit *Hit) *ShotReport {
	timeRequest := this.getDiffSeconds(hit)
	this.MinTime = math.Min(this.MinTime, timeRequest)
	this.MaxTime = math.Max(this.MaxTime, timeRequest)
	this.TotalTime += timeRequest
	this.updateTotalRequests()
	this.updateTotalTransferred(hit)
	this.checkResponseStatusCode(hit)
	this.endTime = hit.EndTime
	return this
}

func (this *ShotReport) GetAvgTime() float64 {
	return (this.MinTime + this.MaxTime) / 2
}

func (this *ShotReport) GetAvailability() float64 {
	return float64(this.CompleteRequests) * 100 / float64(this.TotalRequests)
}

func (this *ShotReport) GetRequestPerSeconds() float64 {
	if this.endTime.Equal(this.startTime) {
		return float64(this.TotalRequests)
	} else {
		return float64(this.TotalRequests) / this.endTime.Sub(this.startTime).Seconds()
	}
}
