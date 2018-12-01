package channelz

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	channelzpb "google.golang.org/grpc/channelz/grpc_channelz_v1"
)

type ChannelzClient struct {
	cc channelzpb.ChannelzClient
}

func NewClient(conn *grpc.ClientConn) *ChannelzClient {
	return &ChannelzClient{
		cc: channelzpb.NewChannelzClient(conn),
	}
}

func (cc *ChannelzClient) DescribeServer(ctx context.Context, name string) {
	server := cc.findServer(ctx, name)
	if server == nil {
		fmt.Printf("server %q not found", name)
		return
	}

	fmt.Printf("Name:\t%s\n", server.Ref.Name)
	fmt.Printf("ServerID:\t%d\n", server.Ref.ServerId)

	fmt.Printf("Calls:\n")
	fmt.Printf("  Started:        \t%d\n", server.Data.CallsStarted)
	fmt.Printf("  Succeeded:      \t%d\n", server.Data.CallsSucceeded)
	fmt.Printf("  Failed:         \t%d\n", server.Data.CallsFailed)
	fmt.Printf("  LastCallStarted:\t%s\n", stringTimestamp(server.Data.LastCallStartedTimestamp))

	if server.Data.Trace != nil {
		fmt.Printf("Trace:\n")
		fmt.Printf("  NumEvents:\t%d\n", server.Data.Trace.NumEventsLogged)
		fmt.Printf("  CreationTimestamp:\t%s\n", stringTimestamp(server.Data.Trace.CreationTimestamp))

		if len(server.Data.Trace.Events) != 0 {
			fmt.Printf("  Events\n")
			fmt.Printf("    %s\t%-80s\t%s\n", "Severity", "Description", "Timestamp")
			for _, ev := range server.Data.Trace.Events {
				fmt.Printf("    %s\t%-80s\t%s\n",
					prettyChannelTraceEventSeverity(ev.Severity), ev.Description, stringTimestamp(ev.Timestamp))
			}
		}
	}
}

func (cc *ChannelzClient) findServer(ctx context.Context, name string) *channelzpb.Server {
	n, err := strconv.Atoi(name)
	if err != nil {
		return cc.findServerByName(ctx, name)
	}
	return cc.findServerByID(ctx, int64(n))
}

func (cc *ChannelzClient) findServerByName(ctx context.Context, name string) *channelzpb.Server {
	var found *channelzpb.Server
	cc.visitGetServers(ctx, func(server *channelzpb.Server) {
		if server.Ref.Name == name {
			if found != nil {
				found = server
			}
		}
	})

	return found
}

func (cc *ChannelzClient) findServerByID(ctx context.Context, id int64) *channelzpb.Server {
	var found *channelzpb.Server
	cc.visitGetServers(ctx, func(server *channelzpb.Server) {
		if server.Ref.ServerId == id {
			found = server
		}
	})

	return found
}

func (cc *ChannelzClient) DescribeChannel(ctx context.Context, name string) {
	channel := cc.findTopChannel(ctx, name)
	if channel == nil {
		fmt.Printf("channel %q not found", name)
		return
	}

	fmt.Printf("Name:     \t%s\n", channel.Ref.Name)
	fmt.Printf("ChannelID:\t%d\n", channel.Ref.ChannelId)
	fmt.Printf("State:    \t%s\n", channel.Data.State.State.String())
	fmt.Printf("Target:   \t%s\n", channel.Data.Target)

	fmt.Printf("Calls:\n")
	fmt.Printf("  Started:        \t%d\n", channel.Data.CallsStarted)
	fmt.Printf("  Succeeded:      \t%d\n", channel.Data.CallsSucceeded)
	fmt.Printf("  Failed:         \t%d\n", channel.Data.CallsFailed)
	fmt.Printf("  LastCallStarted:\t%s\n", stringTimestamp(channel.Data.LastCallStartedTimestamp))

	if len(channel.SocketRef) == 0 {
		fmt.Printf("Socket:   \t%s\n", "<none>")
	} else {
		fmt.Printf("  Sockets\n")
		fmt.Printf("    %s\t%s\n", "SocketID", "Name")
		for _, socket := range channel.SocketRef {
			fmt.Printf("    %s\t%s\t\n", socket.SocketId, socket.Name)
		}
	}

	if len(channel.ChannelRef) == 0 {
		fmt.Printf("Channels:   \t%s\n", "<none>")
	} else {
		fmt.Printf("Channels\n")
		fmt.Printf("  %s\t%s\n", "SocketID", "Name")
		for _, channel := range channel.ChannelRef {
			fmt.Printf("  %d\t%s\n", channel.ChannelId, channel.Name)
		}
	}

	if len(channel.SubchannelRef) == 0 {
		fmt.Printf("Subchannels:   \t%s\n", "<none>")
	} else {
		fmt.Printf("Subchannels:\n")
		fmt.Printf("  %s\t%s\t%s\t%s\t%s\t%s\n", "ID", "Name", "State", "Start", "Succeeded", "Failed")
		for _, subchref := range channel.SubchannelRef {
			res, err := cc.cc.GetSubchannel(ctx, &channelzpb.GetSubchannelRequest{SubchannelId: subchref.SubchannelId})
			if err != nil {
				log.Fatalf("err %v", err)
			}

			subch := res.Subchannel
			fmt.Printf("  %d\t%s\t%s\t%d\t%d\t%d\n",
				subch.Ref.SubchannelId, subch.Ref.Name, subch.Data.State.State.String(),
				subch.Data.CallsStarted,
				subch.Data.CallsSucceeded,
				subch.Data.CallsFailed,
			)
		}
	}

	if channel.Data.Trace != nil {
		fmt.Printf("Trace:\n")
		fmt.Printf("  NumEvents:\t%d\n", channel.Data.Trace.NumEventsLogged)
		fmt.Printf("  CreationTimestamp:\t%s\n", stringTimestamp(channel.Data.Trace.CreationTimestamp))

		if len(channel.Data.Trace.Events) != 0 {
			fmt.Printf("  Events\n")
			fmt.Printf("    %s\t%-80s\t%s\n", "Severity", "Description", "Timestamp")
			for _, ev := range channel.Data.Trace.Events {
				fmt.Printf("    %s\t%-80s\t%s\n",
					prettyChannelTraceEventSeverity(ev.Severity), ev.Description, stringTimestamp(ev.Timestamp))
			}
		}
	}
}

func (cc *ChannelzClient) GetServers(ctx context.Context) {
	now := time.Now()
	cc.visitGetServers(ctx, func(server *channelzpb.Server) {
		fmt.Printf("ID: %v, Name: %v\n", server.Ref.ServerId, server.Ref.Name)

		elapesed := elapsedTimestamp(now, server.Data.LastCallStartedTimestamp)
		fmt.Printf("    [Calls]: Started:%v Succeeded:%v, Failed:%v, Last:%s\n", server.Data.CallsStarted, server.Data.CallsSucceeded, server.Data.CallsFailed, elapesed)

		for _, socket := range server.ListenSocket {
			res, err := cc.cc.GetSocket(ctx, &channelzpb.GetSocketRequest{SocketId: socket.SocketId})
			if err != nil {
				log.Fatalf("err %v\n", err)
			}

			socket := res.Socket
			if socket == nil {
				fmt.Printf("not found\n")
				continue
			}
			fmt.Printf("    [Socket] ID:%v, Name:%v, RemoteName:%v", socket.Ref.SocketId, socket.Ref.Name, socket.RemoteName)
			if addr := socket.Local.GetTcpipAddress(); addr != nil {
				fmt.Printf(", Local IP:%v, Port:%v", net.IP(addr.IpAddress).String(), addr.Port)
			}
			fmt.Println("")
		}

		fmt.Println("")
	})
}

func (cc *ChannelzClient) visitGetServers(ctx context.Context, fn func(*channelzpb.Server)) {
	lastServerID := int64(0)
	for {
		res, err := cc.cc.GetServers(ctx, &channelzpb.GetServersRequest{StartServerId: lastServerID})
		if err != nil {
			log.Fatalf("err: %v\n", err)
		}

		for _, server := range res.Server {
			fn(server)
		}
		if res.End {
			break
		}

		lastServerID++
	}
}

func (cc *ChannelzClient) GetTopChannels(ctx context.Context) {
	now := time.Now()

	cc.visitTopChannels(ctx, func(channel *channelzpb.Channel) {
		fmt.Printf("%s (ID:%d) [%s]\n",
			channel.Data.Target, channel.Ref.ChannelId,
			channel.Data.State.State.String())
		// fmt.Printf("ID: %v, Name: %v\n", channel.Ref.ChannelId, channel.Ref.Name)
		// fmt.Printf("state: %v, Target: %v\n", channel.Data.State.State.String(), channel.Data.Target)

		elapesed := elapsedTimestamp(now, channel.Data.LastCallStartedTimestamp)
		fmt.Printf("  [Calls] Started:%v, Succeeded:%v, Failed:%v, Last:%v\n", channel.Data.CallsStarted, channel.Data.CallsSucceeded, channel.Data.CallsFailed, elapesed)

		// for _, ev := range channel.Data.Trace.Events {
		// 	fmt.Printf("ev %v\n", ev)
		// }

		for _, socket := range channel.SocketRef {
			fmt.Printf("socket %v\n", socket)
		}

		for _, ch := range channel.ChannelRef {
			fmt.Printf("ch %v\n", ch)
		}

		if len(channel.SubchannelRef) != 0 {
			fmt.Printf("  [Subchannels]\n")
		}
		for _, ch := range channel.SubchannelRef {
			res, err := cc.cc.GetSubchannel(ctx, &channelzpb.GetSubchannelRequest{SubchannelId: ch.SubchannelId})
			if err != nil {
				log.Fatalf("err %v", err)
			}

			subch := res.Subchannel
			fmt.Printf("    |-- %s (ID:%d) [%s]\n",
				subch.Data.Target, subch.Ref.SubchannelId,
				subch.Data.State.State.String())

			elapesed := elapsedTimestamp(now, subch.Data.LastCallStartedTimestamp)
			fmt.Printf("          [Calls]: Started:%v, Succeeded:%v, Failed:%v, Last:%s\n", subch.Data.CallsStarted, subch.Data.CallsSucceeded, subch.Data.CallsFailed, elapesed)

			for _, socket := range subch.SocketRef {
				res, err := cc.cc.GetSocket(ctx, &channelzpb.GetSocketRequest{SocketId: socket.SocketId})
				if err != nil {
					log.Fatalf("err %v\n", err)
				}

				socket := res.Socket
				fmt.Printf("          [Socket] ID:%v, Name:%v, RemoteName:%v", socket.Ref.SocketId, socket.Ref.Name, socket.RemoteName)
				fmt.Printf(", Local:")
				if addr := socket.Local.GetTcpipAddress(); addr != nil {
					fmt.Printf("[%v]:%v", net.IP(addr.IpAddress).String(), addr.Port)
				}
				fmt.Printf(" Remote:")
				if addr := socket.Remote.GetTcpipAddress(); addr != nil {
					fmt.Printf("[%v]:%v", net.IP(addr.IpAddress).String(), addr.Port)
				}
				fmt.Printf("\n")
			}

			for _, ch := range subch.ChannelRef {
				fmt.Printf("---- ch %v\n", ch)
			}
			for _, ch := range subch.SubchannelRef {
				fmt.Printf("---- ch %v\n", ch)
			}
		}

		fmt.Println()
	})
}

func (cc *ChannelzClient) findTopChannel(ctx context.Context, name string) *channelzpb.Channel {
	n, err := strconv.Atoi(name)
	if err != nil {
		return cc.findTopChannelByName(ctx, name)
	}
	return cc.findTopChannelByID(ctx, int64(n))
}

func (cc *ChannelzClient) findTopChannelByName(ctx context.Context, name string) *channelzpb.Channel {
	var found *channelzpb.Channel
	cc.visitTopChannels(ctx, func(channel *channelzpb.Channel) {
		if channel.Ref.Name == name {
			if found != nil {
				found = channel
			}
		}
	})

	return found
}

func (cc *ChannelzClient) findTopChannelByID(ctx context.Context, id int64) *channelzpb.Channel {
	var found *channelzpb.Channel
	cc.visitTopChannels(ctx, func(channel *channelzpb.Channel) {
		if channel.Ref.ChannelId == id {
			found = channel
		}
	})

	return found
}

func (cc *ChannelzClient) visitTopChannels(ctx context.Context, fn func(*channelzpb.Channel)) {
	lastChannelID := int64(0)
	retry := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(retry+1))
		defer cancel()
		res, err := cc.cc.GetTopChannels(ctx, &channelzpb.GetTopChannelsRequest{StartChannelId: lastChannelID})
		if err != nil {
			retry++
			continue
		}

		for _, channel := range res.Channel {
			fn(channel)
		}
		if res.End {
			break
		}

		lastChannelID++
	}
}
