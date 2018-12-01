package channelz

import (
	"strings"

	channelzpb "google.golang.org/grpc/channelz/grpc_channelz_v1"
)

func computeIndent(depth int) string {
	return strings.Repeat(" ", depth*2)
}

func prettyChannelTraceEventSeverity(s channelzpb.ChannelTraceEvent_Severity) string {
	return strings.TrimPrefix(s.String(), "CT_")
}
