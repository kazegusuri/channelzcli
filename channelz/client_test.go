package channelz

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func newTestClient1(b *bytes.Buffer) *ChannelzClient {
	return &ChannelzClient{
		w:  b,
		cc: fakeChannelzClient1,
	}
}

func assertOutput(t *testing.T, expected, actual string) {
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)
	if expected != actual {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expected, actual)
	}
}

func TestDescribeServer(t *testing.T) {
	b := &bytes.Buffer{}
	ctx := context.Background()
	c := newTestClient1(b)

	t.Run("ByID", func(t *testing.T) {
		t.Run("server0", func(t *testing.T) {
			b.Reset()
			c.DescribeServer(ctx, "0")
			expected := `
Name:	server0
ServerID:	0
Calls:
  Started:        	100
  Succeeded:      	90
  Failed:         	10
  LastCallStarted:	none
`
			assertOutput(t, expected, b.String())
		})

		t.Run("server1", func(t *testing.T) {
			b.Reset()
			c.DescribeServer(ctx, "1")
			expected := `
Name:	server1
ServerID:	1
Calls:
  Started:        	110
  Succeeded:      	99
  Failed:         	11
  LastCallStarted:	2018-12-01 21:33:20.123456789 +0000 UTC
`
			assertOutput(t, expected, b.String())
		})
	})

	t.Run("ByName", func(t *testing.T) {
		t.Run("server0", func(t *testing.T) {
			b.Reset()
			c.DescribeServer(ctx, "server0")
			expected := `
Name:	server0
ServerID:	0
Calls:
  Started:        	100
  Succeeded:      	90
  Failed:         	10
  LastCallStarted:	none
`
			assertOutput(t, expected, b.String())
		})
	})
}
