package channelz

import (
	"strings"

	channelzpb "google.golang.org/grpc/channelz/grpc_channelz_v1"
)

func decorateEmpty(s string) string {
	if s == "" {
		return "<none>"
	}
	return s
}

func prettyChannelTraceEventSeverity(s channelzpb.ChannelTraceEvent_Severity) string {
	return strings.TrimPrefix(s.String(), "CT_")
}
