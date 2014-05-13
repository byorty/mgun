package target

import (
	"time"
	"net/http"
)

type Hit struct {
	StartTime time.Time
	EndTime time.Time
	Shot *Shot
	Request *http.Request
	Response *http.Response
	ResponseBody []byte
}
