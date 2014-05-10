package gun

import (
	"fmt"
	"mgun/target"
	"time"
	tm "github.com/buger/goterm"
)

func NewReporter() *Reporter {
	fmt.Println("")
	reporter := new(Reporter)
	reporter.startTime = time.Now()
	reporter.reports = make(map[string]*target.ShotReport)
	return reporter
}

type Reporter struct {
	startTime time.Time
	reports map[string]*target.ShotReport
}

func (this *Reporter) report(newTarget *target.Target, hits <- chan *target.Hit) {
	hitsTable := tm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(hitsTable, "Path\tComplete\tFailed\tMin\tMax\tAvg\tAvailability\tRequests per sec.\n")
	for hit := range hits {
		path := hit.Shot.GetPath()
		if report, ok := this.reports[path]; ok {
			report.Update(hit)
		} else {
			report := target.NewStageReport(hit)
			this.reports[path] = report
		}
	}

	reportsCount := float64(len(this.reports))
	var totalRequests int
	var completeRequests int
	var failedRequests int
	var availability float64
	var requestPerSeconds float64

	for path, report := range this.reports {

		totalRequests += report.TotalRequests
		completeRequests += report.CompleteRequests
		failedRequests += report.FailedRequests
		availability += report.GetAvailability()
		requestPerSeconds += report.GetRequestPerSeconds()

		fmt.Fprintf(hitsTable, "%s\t%d\t%d\t%.3fs.\t%.3fs.\t%.3fs.\t%.2f%%\t~ %.2f\n",
			path,
			report.CompleteRequests,
			report.FailedRequests,
			report.MinTime,
			report.MaxTime,
			report.GetAvgTime(),
			report.GetAvailability(),
			report.GetRequestPerSeconds(),
		)
	}

	targetTable := tm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(targetTable, "Server Hostname:\t%s\n", newTarget.Host)
	fmt.Fprintf(targetTable, "Server Port:\t%d\n", newTarget.Port)
	fmt.Fprintf(targetTable, "Concurrency Level:\t%d\n", newTarget.Concurrency)
	fmt.Fprintf(targetTable, "Time taken for tests:\t%.3f seconds\n", time.Now().Sub(this.startTime).Seconds())
	fmt.Fprintf(targetTable, "Total requests:\t%d\n", totalRequests)
	fmt.Fprintf(targetTable, "Complete requests:\t%d\n", completeRequests)
	fmt.Fprintf(targetTable, "Failed requests:\t%d\n", failedRequests)
	fmt.Fprintf(targetTable, "Availability:\t%.2f%%\n", availability / reportsCount)
	fmt.Fprintf(targetTable, "Requests per second:\t~ %.2f\n", requestPerSeconds / reportsCount)

	fmt.Println("")
	fmt.Println("")
	fmt.Println(targetTable)
	fmt.Println(hitsTable)
}

