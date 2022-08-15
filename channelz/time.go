package channelz

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func stringTimestamp(ts *timestamppb.Timestamp) string {
	if ts != nil && ts.Seconds == 0 && ts.Nanos == 0 {
		return "none"
	}

	return ts.AsTime().UTC().String()
}

func elapsedTimestamp(now time.Time, ts *timestamppb.Timestamp) string {
	if ts != nil && ts.Seconds == 0 && ts.Nanos == 0 {
		return "none"
	}

	return prettyDuration(now.Sub(ts.AsTime()))
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
