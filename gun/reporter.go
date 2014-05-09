package gun

import (
	"fmt"
	"mgun/target"
	"github.com/buger/goterm"
)

func NewReporter() *Reporter {
	reporter := new(Reporter)
	reporter.reports = make(map[string]*target.ShotReport)
	return reporter
}

type Reporter struct {
	reports map[string]*target.ShotReport
}

func (this *Reporter) report(hits <- chan *target.Hit) {
	table := goterm.NewTable(0, 0, 2, ' ', 0)
	fmt.Fprintf(table, "Path\tComplete\tFailed\tMin\tMax\tAvg\tAvailability\tRequests per second\n")
	for hit := range hits {
		path := hit.Shot.GetPath()
		if report, ok := this.reports[path]; ok {
			report.Update(hit)
		} else {
			report := target.NewStageReport(hit)
			this.reports[path] = report
		}
	}
	for path, repoprt := range this.reports {
		fmt.Fprintf(table, "%s\t%d\t%d\t%.3fs.\t%.3fs.\t%.3fs.\t%.2f%%\t%.2f\n",
			path,
			repoprt.CompleteRequests,
			repoprt.FailedRequests,
			repoprt.MinTime,
			repoprt.MaxTime,
			repoprt.GetAvgTime(),
			repoprt.GetAvailability(),
			repoprt.GetRequestPerSeconds(),
		)
	}
	fmt.Println("\n", table)
}

