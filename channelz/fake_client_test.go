package channelz

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	channelzpb "google.golang.org/grpc/channelz/grpc_channelz_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ channelzpb.ChannelzClient = (*fakeChannelzClient)(nil)

type fakeChannelzClient struct {
	topChannels []*channelzpb.Channel
	servers     []*channelzpb.Server
	channels    []*channelzpb.Channel
	subchannels []*channelzpb.Subchannel
	sockets     []*channelzpb.Socket
}

func (c *fakeChannelzClient) GetTopChannels(ctx context.Context, in *channelzpb.GetTopChannelsRequest, opts ...grpc.CallOption) (*channelzpb.GetTopChannelsResponse, error) {
	return &channelzpb.GetTopChannelsResponse{
		Channel: c.topChannels,
		End:     true,
	}, nil
}

func (c *fakeChannelzClient) GetServers(ctx context.Context, in *channelzpb.GetServersRequest, opts ...grpc.CallOption) (*channelzpb.GetServersResponse, error) {
	return &channelzpb.GetServersResponse{
		Server: c.servers,
		End:    true,
	}, nil
}

func (c *fakeChannelzClient) GetServer(ctx context.Context, in *channelzpb.GetServerRequest, opts ...grpc.CallOption) (*channelzpb.GetServerResponse, error) {
	for _, s := range c.servers {
		if in.ServerId == s.Ref.ServerId {
			return &channelzpb.GetServerResponse{
				Server: s,
			}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "not found")
}

func (c *fakeChannelzClient) GetServerSockets(ctx context.Context, in *channelzpb.GetServerSocketsRequest, opts ...grpc.CallOption) (*channelzpb.GetServerSocketsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (c *fakeChannelzClient) GetChannel(ctx context.Context, in *channelzpb.GetChannelRequest, opts ...grpc.CallOption) (*channelzpb.GetChannelResponse, error) {
	for _, ch := range c.channels {
		if in.ChannelId == ch.Ref.ChannelId {
			return &channelzpb.GetChannelResponse{
				Channel: ch,
			}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "not found")
}

func (c *fakeChannelzClient) GetSubchannel(ctx context.Context, in *channelzpb.GetSubchannelRequest, opts ...grpc.CallOption) (*channelzpb.GetSubchannelResponse, error) {
	for _, ch := range c.subchannels {
		if in.SubchannelId == ch.Ref.SubchannelId {
			return &channelzpb.GetSubchannelResponse{
				Subchannel: ch,
			}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "not found")
}

func (c *fakeChannelzClient) GetSocket(ctx context.Context, in *channelzpb.GetSocketRequest, opts ...grpc.CallOption) (*channelzpb.GetSocketResponse, error) {
	for _, socket := range c.sockets {
		if in.SocketId == socket.Ref.SocketId {
			return &channelzpb.GetSocketResponse{
				Socket: socket,
			}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "not found")
}

var (
	channelCount    int64
	subchannelCount int64
	socketCount     int64
	serverCount     int64
)

type channelParam struct {
	state                    channelzpb.ChannelConnectivityState_State
	lastCallStartedTimestamp *timestamppb.Timestamp
	chRef                    []*channelzpb.ChannelRef
	subchRef                 []*channelzpb.SubchannelRef
	sockRef                  []*channelzpb.SocketRef
}

type socketParam struct {
	remoteIP   net.IP
	remotePort int32
	localIP    net.IP
	localPort  int32
}

type serverParam struct {
	lastCallStartedTimestamp *timestamppb.Timestamp
	sockRef                  []*channelzpb.SocketRef
}

func testCreateChannel(param channelParam) *channelzpb.Channel {
	id := channelCount
	channelCount++
	return &channelzpb.Channel{
		Ref: &channelzpb.ChannelRef{
			ChannelId: id,
			Name:      fmt.Sprintf("foo%d", id),
		},
		Data: &channelzpb.ChannelData{
			State: &channelzpb.ChannelConnectivityState{
				State: param.state,
			},
			Target:                   fmt.Sprintf("foo%d.test.com", id),
			Trace:                    &channelzpb.ChannelTrace{},
			CallsStarted:             100 + id*10,
			CallsSucceeded:           90 + id*9,
			CallsFailed:              10 + id,
			LastCallStartedTimestamp: param.lastCallStartedTimestamp,
		},
		ChannelRef:    param.chRef,
		SubchannelRef: param.subchRef,
		SocketRef:     param.sockRef,
	}
}

func testCreateSubchannel(param channelParam) *channelzpb.Subchannel {
	id := subchannelCount
	subchannelCount++
	return &channelzpb.Subchannel{
		Ref: &channelzpb.SubchannelRef{
			SubchannelId: id,
			Name:         fmt.Sprintf("bar%d", id),
		},
		Data: &channelzpb.ChannelData{
			State: &channelzpb.ChannelConnectivityState{
				State: param.state,
			},
			Target:                   fmt.Sprintf("bar%d.test.com", id),
			Trace:                    &channelzpb.ChannelTrace{},
			CallsStarted:             100 + id*10,
			CallsSucceeded:           90 + id*9,
			CallsFailed:              10 + id,
			LastCallStartedTimestamp: param.lastCallStartedTimestamp,
		},
		ChannelRef:    param.chRef,
		SubchannelRef: param.subchRef,
		SocketRef:     param.sockRef,
	}
}

func testCreateSocket(param socketParam) *channelzpb.Socket {
	id := socketCount
	socketCount++

	var localAddr *channelzpb.Address
	var remoteAddr *channelzpb.Address

	if param.localIP != nil {
		localAddr = &channelzpb.Address{
			Address: &channelzpb.Address_TcpipAddress{
				TcpipAddress: &channelzpb.Address_TcpIpAddress{
					IpAddress: []byte(param.localIP),
					Port:      param.localPort,
				},
			},
		}
	}

	if param.remoteIP != nil {
		remoteAddr = &channelzpb.Address{
			Address: &channelzpb.Address_TcpipAddress{
				TcpipAddress: &channelzpb.Address_TcpIpAddress{
					IpAddress: []byte(param.remoteIP),
					Port:      param.remotePort,
				},
			},
		}
	}

	return &channelzpb.Socket{
		Ref: &channelzpb.SocketRef{
			SocketId: id,
			Name:     fmt.Sprintf("sock%d", id),
		},
		Data:       &channelzpb.SocketData{},
		Local:      localAddr,
		Remote:     remoteAddr,
		Security:   &channelzpb.Security{},
		RemoteName: "",
	}
}

func testCreateServer(param serverParam) *channelzpb.Server {
	id := serverCount
	serverCount++

	return &channelzpb.Server{
		Ref: &channelzpb.ServerRef{
			ServerId: id,
			Name:     fmt.Sprintf("server%d", id),
		},
		Data: &channelzpb.ServerData{
			CallsStarted:             100 + id*10,
			CallsSucceeded:           90 + id*9,
			CallsFailed:              10 + id,
			LastCallStartedTimestamp: param.lastCallStartedTimestamp,
		},
		ListenSocket: param.sockRef,
	}
}

var (
	fakeChannelzClient1 channelzpb.ChannelzClient
)

func init() {
	now := fixedTime
	ts1 := timestamppb.New(now)

	srvsock1 := testCreateSocket(socketParam{
		localIP:   net.IPv4(127, 0, 1, 2),
		localPort: 9000,
	})
	srvsock2 := testCreateSocket(socketParam{
		localIP:   net.IPv4(127, 0, 1, 2),
		localPort: 9001,
	})
	srv1 := testCreateServer(serverParam{
		sockRef: []*channelzpb.SocketRef{srvsock1.Ref},
	})
	srv2 := testCreateServer(serverParam{
		lastCallStartedTimestamp: ts1,
		sockRef:                  []*channelzpb.SocketRef{srvsock2.Ref},
	})

	subchsock1 := testCreateSocket(socketParam{
		localIP:    net.IPv4(127, 0, 1, 2),
		localPort:  9001,
		remoteIP:   net.IPv4(111, 111, 111, 111),
		remotePort: 30000,
	})
	subch1 := testCreateSubchannel(channelParam{
		state:                    channelzpb.ChannelConnectivityState_READY,
		lastCallStartedTimestamp: ts1,
		sockRef:                  []*channelzpb.SocketRef{subchsock1.Ref},
	})
	topch1 := testCreateChannel(channelParam{
		state:                    channelzpb.ChannelConnectivityState_READY,
		lastCallStartedTimestamp: ts1,
		subchRef:                 []*channelzpb.SubchannelRef{subch1.Ref},
	})

	var subchSocks1 []*channelzpb.Socket
	var subchs1 []*channelzpb.Subchannel
	var subchRef1 []*channelzpb.SubchannelRef
	for i := 0; i < 4; i++ {
		socket := testCreateSocket(socketParam{
			localIP:    net.IPv4(127, 0, 1, 2),
			localPort:  9001,
			remoteIP:   net.IPv4(111, 111, 111, byte(112+i)),
			remotePort: int32(30001 + i),
		})
		subchSocks1 = append(subchSocks1, socket)
		subch := testCreateSubchannel(channelParam{
			state:                    channelzpb.ChannelConnectivityState_READY,
			lastCallStartedTimestamp: ts1,
			sockRef:                  []*channelzpb.SocketRef{socket.Ref},
		})
		subchs1 = append(subchs1, subch)
		subchRef1 = append(subchRef1, subch.Ref)
	}
	topch2 := testCreateChannel(channelParam{
		state:                    channelzpb.ChannelConnectivityState_READY,
		lastCallStartedTimestamp: ts1,
		subchRef:                 subchRef1,
	})

	fakeChannelzClient1 = &fakeChannelzClient{
		topChannels: []*channelzpb.Channel{topch1, topch2},
		channels: []*channelzpb.Channel{
			topch1, topch2,
		},
		subchannels: append([]*channelzpb.Subchannel{
			subch1,
		}, subchs1...),
		sockets: append([]*channelzpb.Socket{
			srvsock1, srvsock2,
			subchsock1,
		}, subchSocks1...),
		servers: []*channelzpb.Server{srv1, srv2},
	}
}
