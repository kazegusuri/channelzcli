package channelz

import (
	"time"
)

var fixedTime = time.Unix(1543700000, 123456789).UTC()

func init() {
	timeNow = func() time.Time {
		return fixedTime
	}
}
