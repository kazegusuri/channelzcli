package channelz

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func stringTimestamp(ts *timestamp.Timestamp) string {
	if ts != nil && ts.Seconds == 0 && ts.Nanos == 0 {
		return "none"
	}

	pt, err := ptypes.Timestamp(ts)
	if err != nil {
		return "none"
	}

	return pt.UTC().String()
}

func elapsedTimestamp(now time.Time, ts *timestamp.Timestamp) string {
	if ts != nil && ts.Seconds == 0 && ts.Nanos == 0 {
		return "none"
	}

	pt, err := ptypes.Timestamp(ts)
	if err != nil {
		return "none"
	}

	return prettyDuration(now.Sub(pt))
}

func prettyDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	if d > (24 * time.Hour) {
		v := d / (24 * time.Hour)
		return fmt.Sprintf("%dd", v)
	} else if d > time.Hour {
		v := d / time.Hour
		return fmt.Sprintf("%dh", v)
	} else if d > time.Minute {
		v := d / time.Minute
		return fmt.Sprintf("%dm", v)
	} else if d > time.Second {
		v := d / time.Second
		return fmt.Sprintf("%ds", v)
	} else {
		v := d / time.Millisecond
		return fmt.Sprintf("%dms", v)
	}
}
