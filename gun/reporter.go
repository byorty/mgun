package gun

import (
	"fmt"
	"github.com/byorty/mgun/target"
	"time"
	hm "github.com/dustin/go-humanize"
	tm "github.com/buger/goterm"
)

func NewReporter() *Reporter {
	fmt.Println("")
	reporter := new(Reporter)
	reporter.startTime = time.Now()
	return reporter
}

type Reporter struct {
	startTime time.Time
	reports map[string]*target.ShotReport
}

func (this *Reporter) report(newTarget *target.Target, hits <- chan *target.Hit) {
	reports := make(map[string]*target.ShotReport)
	hitsTable := tm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(hitsTable, "Request\tCompl.\tFail.\tMin\tMax\tAvg\tAvail.\tReq. per sec.\tContent len.\tTotal trans.\n")
	for hit := range hits {
		key := this.getRequestName(hit.Shot)
		if report, ok := reports[key]; ok {
			report.Update(hit)
		} else {
			report := target.NewStageReport(hit)
			reports[key] = report
		}
	}

	reportsCount := float64(len(reports))
	var totalRequests int
	var completeRequests int
	var failedRequests int
	var availability float64
	var requestPerSeconds float64
	var totalTransferred int64

	for _, shot := range newTarget.Shots {

		name := this.getRequestName(shot)
		report := reports[name]

		totalRequests += report.TotalRequests
		completeRequests += report.CompleteRequests
		failedRequests += report.FailedRequests
		availability += report.GetAvailability()
		totalTransferred += report.TotalTransferred
		requestPerSeconds += report.GetRequestPerSeconds()

		fmt.Fprintf(hitsTable, "%s\t%d\t%d\t%.3fs.\t%.3fs.\t%.3fs.\t%.2f%%\t~ %.2f\t%s\t%s\n",
			name,
			report.CompleteRequests,
			report.FailedRequests,
			report.MinTime,
			report.MaxTime,
			report.GetAvgTime(),
			report.GetAvailability(),
			report.GetRequestPerSeconds(),
			hm.Bytes(uint64(report.ContentLength)),
			hm.Bytes(uint64(report.TotalTransferred)),
		)
	}

	targetTable := tm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(targetTable, "Server Hostname:\t%s\n", newTarget.Host)
	fmt.Fprintf(targetTable, "Server Port:\t%d\n", newTarget.Port)
	fmt.Fprintf(targetTable, "Concurrency Level:\t%d\n", newTarget.Concurrency)
	fmt.Fprintf(targetTable, "Loop count:\t%d\n", newTarget.LoopCount)
	fmt.Fprintf(targetTable, "Timeout:\t%d seconds\n", newTarget.Timeout)
	fmt.Fprintf(targetTable, "Time taken for tests:\t%.3f seconds\n", time.Now().Sub(this.startTime).Seconds())
	fmt.Fprintf(targetTable, "Total requests:\t%d\n", totalRequests)
	fmt.Fprintf(targetTable, "Complete requests:\t%d\n", completeRequests)
	fmt.Fprintf(targetTable, "Failed requests:\t%d\n", failedRequests)
	fmt.Fprintf(targetTable, "Availability:\t%.2f%%\n", availability / reportsCount)
	fmt.Fprintf(targetTable, "Requests per second:\t~ %.2f\n", requestPerSeconds / reportsCount)
	fmt.Fprintf(targetTable, "Total transferred:\t%s\n", hm.Bytes(uint64(totalTransferred)))

	fmt.Println("")
	fmt.Println("")
	fmt.Println(targetTable)
	fmt.Println(hitsTable)
}

func (this *Reporter) getRequestName(shot *target.Shot) string {
	return fmt.Sprintf("%s %s", shot.GetMethod(), shot.GetPath())
}

