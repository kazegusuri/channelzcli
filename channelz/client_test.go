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

	t.Run("server0", func(t *testing.T) {
		expected := `
Name:	server0
ServerID:	0
Calls:
  Started:        	100
  Succeeded:      	90
  Failed:         	10
  LastCallStarted:	none
`
		t.Run("ByID", func(t *testing.T) {
			b.Reset()
			c.DescribeServer(ctx, "0")
			assertOutput(t, expected, b.String())
		})
		t.Run("ByName", func(t *testing.T) {
			b.Reset()
			c.DescribeServer(ctx, "server0")
			assertOutput(t, expected, b.String())
		})
	})

	t.Run("server1", func(t *testing.T) {
		expected := `
Name:	server1
ServerID:	1
Calls:
  Started:        	110
  Succeeded:      	99
  Failed:         	11
  LastCallStarted:	2018-12-01 21:33:20.123456789 +0000 UTC
`
		t.Run("ByID", func(t *testing.T) {
			b.Reset()
			c.DescribeServer(ctx, "1")
			assertOutput(t, expected, b.String())
		})
		t.Run("ByName", func(t *testing.T) {
			b.Reset()
			c.DescribeServer(ctx, "server1")
			assertOutput(t, expected, b.String())
		})
	})
}

func TestDescribeChannel(t *testing.T) {
	b := &bytes.Buffer{}
	ctx := context.Background()
	c := newTestClient1(b)

	t.Run("TopChannel", func(t *testing.T) {
		expected := `
Name:     	foo0
ChannelID:	0
State:    	READY
Target:   	foo0.test.com
Calls:
  Started:        	100
  Succeeded:      	90
  Failed:         	10
  LastCallStarted:	2018-12-01 21:33:20.123456789 +0000 UTC
Socket:   	<none>
Channels:   	<none>
Subchannels:
  ID	Name	State	Start	Succeeded	Failed
  0	bar0	READY	100	90	10
Trace:
  NumEvents:	0
  CreationTimestamp:	none
`
		t.Run("ByID", func(t *testing.T) {
			b.Reset()
			c.DescribeChannel(ctx, "0")
			assertOutput(t, expected, b.String())
		})
		t.Run("ByName", func(t *testing.T) {
			b.Reset()
			c.DescribeChannel(ctx, "foo0")
			assertOutput(t, expected, b.String())
		})
	})

	t.Run("TopChannelWithSubChannels", func(t *testing.T) {
		expected := `
Name:     	foo1
ChannelID:	1
State:    	READY
Target:   	foo1.test.com
Calls:
  Started:        	110
  Succeeded:      	99
  Failed:         	11
  LastCallStarted:	2018-12-01 21:33:20.123456789 +0000 UTC
Socket:   	<none>
Channels:   	<none>
Subchannels:
  ID	Name	State	Start	Succeeded	Failed
  1	bar1	READY	110	99	11
  2	bar2	READY	120	108	12
  3	bar3	READY	130	117	13
  4	bar4	READY	140	126	14
Trace:
  NumEvents:	0
  CreationTimestamp:	none
`
		t.Run("ByID", func(t *testing.T) {
			b.Reset()
			c.DescribeChannel(ctx, "1")
			assertOutput(t, expected, b.String())
		})
		t.Run("ByName", func(t *testing.T) {
			b.Reset()
			c.DescribeChannel(ctx, "foo1")
			assertOutput(t, expected, b.String())
		})
	})
}

func TestListServers(t *testing.T) {
	b := &bytes.Buffer{}
	ctx := context.Background()
	c := newTestClient1(b)

	t.Run("server", func(t *testing.T) {
		expected := `
ID	Name	LocalAddr	Calls	Success	Fail	LastCall
0	server0	[127.0.1.2]:9000	100   	90    	10    	none
1	server1	[127.0.1.2]:9001	110   	99    	11    	0ms
`
		b.Reset()
		c.ListServers(ctx)
		assertOutput(t, expected, b.String())
	})
}

func TestListChannels(t *testing.T) {
	b := &bytes.Buffer{}
	ctx := context.Background()
	c := newTestClient1(b)

	t.Run("server", func(t *testing.T) {
		expected := `
ID	Name                                                                            	State	Channel	SubChannel	Calls	Success	Fail	LastCall
0	foo0                                                                            	READY	0      	1         	100   	90    	10    	0ms     
1	foo1                                                                            	READY	0      	4         	110   	99    	11    	0ms
`
		b.Reset()
		c.ListTopChannels(ctx)
		assertOutput(t, expected, b.String())
	})
}
