package channelz

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	channelzpb "google.golang.org/grpc/channelz/grpc_channelz_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var timeNow = time.Now

type ChannelzClient struct {
	cc channelzpb.ChannelzClient
	w  io.Writer
}

func NewClient(conn *grpc.ClientConn, w io.Writer) *ChannelzClient {
	return &ChannelzClient{
		cc: channelzpb.NewChannelzClient(conn),
		w:  w,
	}
}

func (cc *ChannelzClient) printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(cc.w, format, a...)
}

func (cc *ChannelzClient) DescribeServer(ctx context.Context, name string) {
	server := cc.findServer(ctx, name)
	if server == nil {
		fmt.Printf("server %q not found", name)
		return
	}

	cc.printf("ID: \t%d\n", server.Ref.ServerId)
	cc.printf("Name:\t%s\n", server.Ref.Name)

	cc.printf("Calls:\n")
	cc.printf("  Started:        \t%d\n", server.Data.CallsStarted)
	cc.printf("  Succeeded:      \t%d\n", server.Data.CallsSucceeded)
	cc.printf("  Failed:         \t%d\n", server.Data.CallsFailed)
	cc.printf("  LastCallStarted:\t%s\n", stringTimestamp(server.Data.LastCallStartedTimestamp))

	if server.Data.Trace != nil {
		cc.printf("Trace:\n")
		cc.printf("  NumEvents:\t%d\n", server.Data.Trace.NumEventsLogged)
		cc.printf("  CreationTimestamp:\t%s\n", stringTimestamp(server.Data.Trace.CreationTimestamp))

		if len(server.Data.Trace.Events) != 0 {
			cc.printf("  Events\n")
			cc.printf("    %s\t%-80s\t%s\n", "Severity", "Description", "Timestamp")
			for _, ev := range server.Data.Trace.Events {
				cc.printf("    %s\t%-80s\t%s\n",
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
			if found == nil {
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
		cc.printf("channel %q not found", name)
		return
	}

	cc.printf("ID:       \t%d\n", channel.Ref.ChannelId)
	cc.printf("Name:     \t%s\n", channel.Ref.Name)
	cc.printf("State:    \t%s\n", channel.Data.State.State.String())
	cc.printf("Target:   \t%s\n", channel.Data.Target)

	cc.printf("Calls:\n")
	cc.printf("  Started:    \t%d\n", channel.Data.CallsStarted)
	cc.printf("  Succeeded:  \t%d\n", channel.Data.CallsSucceeded)
	cc.printf("  Failed:     \t%d\n", channel.Data.CallsFailed)
	cc.printf("  LastCallStarted:\t%s\n", stringTimestamp(channel.Data.LastCallStartedTimestamp))

	if len(channel.SocketRef) == 0 {
		cc.printf("Socket:   \t%s\n", "<none>")
	} else {
		cc.printf("  Sockets\n")
		cc.printf("    %s\t%s\n", "SocketID", "Name")
		for _, socket := range channel.SocketRef {
			cc.printf("    %d\t%s\t\n", socket.SocketId, socket.Name)
		}
	}

	if len(channel.ChannelRef) == 0 {
		cc.printf("Channels:   \t%s\n", "<none>")
	} else {
		cc.printf("Channels\n")
		cc.printf("  %s\t%s\n", "SocketID", "Name")
		for _, channel := range channel.ChannelRef {
			cc.printf("  %d\t%s\n", channel.ChannelId, channel.Name)
		}
	}

	if len(channel.SubchannelRef) == 0 {
		cc.printf("Subchannels:   \t%s\n", "<none>")
	} else {
		cc.printf("Subchannels:\n")
		cc.printf("  %s\t%s\t%s\t%-6s\t%-8s\t%-6s\n", "ID", "Name", "State", "Start", "Succeeded", "Failed")
		for _, subchref := range channel.SubchannelRef {
			res, err := cc.cc.GetSubchannel(ctx, &channelzpb.GetSubchannelRequest{SubchannelId: subchref.SubchannelId})
			if err != nil {
				log.Fatalf("err %v", err)
			}

			subch := res.Subchannel
			cc.printf("  %d\t%s\t%s\t%-6d\t%-8d\t%-6d\n",
				subch.Ref.SubchannelId, subch.Ref.Name, subch.Data.State.State.String(),
				subch.Data.CallsStarted,
				subch.Data.CallsSucceeded,
				subch.Data.CallsFailed,
			)
		}
	}

	if channel.Data.Trace != nil {
		cc.printf("Trace:\n")
		cc.printf("  NumEvents:\t%d\n", channel.Data.Trace.NumEventsLogged)
		cc.printf("  CreationTimestamp:\t%s\n", stringTimestamp(channel.Data.Trace.CreationTimestamp))

		if len(channel.Data.Trace.Events) != 0 {
			cc.printf("  Events\n")
			cc.printf("    %s\t%-80s\t%s\n", "Severity", "Description", "Timestamp")
			for _, ev := range channel.Data.Trace.Events {
				cc.printf("    %s\t%-80s\t%s\n",
					prettyChannelTraceEventSeverity(ev.Severity), ev.Description, stringTimestamp(ev.Timestamp))
			}
		}
	}
}

func (cc *ChannelzClient) findSocketByID(ctx context.Context, id int64) *channelzpb.Socket {
	res, err := cc.cc.GetSocket(ctx, &channelzpb.GetSocketRequest{SocketId: id})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}
		log.Fatalf("err: %v\n", err)
	}

	return res.Socket
}

func (cc *ChannelzClient) DescribeServerSocket(ctx context.Context, name string) {
	id, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		// TODO: find by name
		cc.printf("serversocket %q not found", name)
		return
	}

	socket := cc.findSocketByID(ctx, id)
	if socket == nil {
		cc.printf("serversocket %q not found", name)
		return
	}

	cc.printf("ID:       \t%d\n", socket.Ref.SocketId)
	cc.printf("Name:     \t%s\n", socket.Ref.Name)
	cc.printf("Local:    \t%s\n", addrToString(socket.Local))
	cc.printf("Remote:   \t%s\n", addrToString(socket.Remote))

	cc.printf("Streams:\n")
	cc.printf("  Started:    \t%d\n", socket.Data.StreamsStarted)
	cc.printf("  Succeeded:  \t%d\n", socket.Data.StreamsSucceeded)
	cc.printf("  Failed:     \t%d\n", socket.Data.StreamsFailed)
	cc.printf("  LastCreated:\t%s\n", stringTimestamp(socket.Data.LastRemoteStreamCreatedTimestamp))

	cc.printf("Messages:\n")
	cc.printf("  Sent:    \t%d\n", socket.Data.MessagesSent)
	cc.printf("  Recieved:  \t%d\n", socket.Data.MessagesReceived)
	cc.printf("  LastSent:\t%s\n", stringTimestamp(socket.Data.LastMessageSentTimestamp))
	cc.printf("  LastReceived:\t%s\n", stringTimestamp(socket.Data.LastMessageReceivedTimestamp))

	cc.printf("Options:\n")
	for _, opt := range socket.Data.Option {
		cc.printf("  %s:\t%s\n", opt.Name, opt.Value)
	}

	cc.printf("Security:\n")
	if socket.Security == nil {
		cc.printf("  Model: none\n")
	} else {

		switch socket.Security.GetModel().(type) {
		case *channelzpb.Security_Tls_:
			cc.printf("  Model: tls\n")
		case *channelzpb.Security_Other:
			cc.printf("  Model: other\n")
		}
	}
}

func (cc *ChannelzClient) ListServers(ctx context.Context) {
	now := timeNow()

	cc.printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		"ID", "Name", "LocalAddr", "Calls", "Success", "Fail", "LastCall")

	cc.visitGetServers(ctx, func(server *channelzpb.Server) {
		// see first socket only
		var socket *channelzpb.Socket
		if len(server.ListenSocket) > 0 {
			res, err := cc.cc.GetSocket(ctx, &channelzpb.GetSocketRequest{SocketId: server.ListenSocket[0].SocketId})
			if err != nil {
				log.Fatalf("err %v\n", err)
			}
			socket = res.Socket
		}

		var localAddr string
		if addr := socket.GetLocal().GetTcpipAddress(); addr != nil {
			localAddr = fmt.Sprintf("[%v]:%v", net.IP(addr.IpAddress).String(), addr.Port)
		}

		cc.printf("%d\t%s\t%-12s\t%-6d\t%-6d\t%-6d\t%s\n",
			server.Ref.ServerId,
			decorateEmpty(server.Ref.Name),
			decorateEmpty(localAddr),
			server.Data.CallsStarted,
			server.Data.CallsSucceeded,
			server.Data.CallsFailed,
			elapsedTimestamp(now, server.Data.LastCallStartedTimestamp),
		)
	})
}

func (cc *ChannelzClient) TreeServers(ctx context.Context) {
	now := timeNow()
	cc.visitGetServers(ctx, func(server *channelzpb.Server) {
		cc.printf("ID: %v, Name: %v\n", server.Ref.ServerId, server.Ref.Name)

		elapesed := elapsedTimestamp(now, server.Data.LastCallStartedTimestamp)
		cc.printf("    [Calls]: Started:%v Succeeded:%v, Failed:%v, Last:%s\n", server.Data.CallsStarted, server.Data.CallsSucceeded, server.Data.CallsFailed, elapesed)

		for _, socket := range server.ListenSocket {
			res, err := cc.cc.GetSocket(ctx, &channelzpb.GetSocketRequest{SocketId: socket.SocketId})
			if err != nil {
				log.Fatalf("err %v\n", err)
			}

			socket := res.Socket
			if socket == nil {
				cc.printf("not found\n")
				continue
			}
			cc.printf("    [Socket] ID:%v, Name:%v, RemoteName:%v", socket.Ref.SocketId, socket.Ref.Name, socket.RemoteName)
			if addr := socket.Local.GetTcpipAddress(); addr != nil {
				cc.printf(", Local IP:%v, Port:%v", net.IP(addr.IpAddress).String(), addr.Port)
			}
			cc.printf("\n")
		}

		cc.printf("\n")
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

func (cc *ChannelzClient) ListTopChannels(ctx context.Context) {
	now := timeNow()

	cc.printf("%s\t%-80s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		"ID", "Name", "State", "Channel", "SubChannel", "Calls", "Success", "Fail", "LastCall")

	cc.visitTopChannels(ctx, func(channel *channelzpb.Channel) {
		cc.printf("%d\t%-80s\t%s\t%-7d\t%-10d\t%-6d\t%-6d\t%-6d\t%-8s\n",
			channel.Ref.ChannelId,
			decorateEmpty(channel.Ref.Name),
			channel.Data.State.State.String(),
			len(channel.ChannelRef),
			len(channel.SubchannelRef),
			channel.Data.CallsStarted,
			channel.Data.CallsSucceeded,
			channel.Data.CallsFailed,
			elapsedTimestamp(now, channel.Data.LastCallStartedTimestamp),
		)
	})
}

func addrToString(addr *channelzpb.Address) string {
	if tcpaddr := addr.GetTcpipAddress(); tcpaddr != nil {
		return fmt.Sprintf("[%v]:%v", net.IP(tcpaddr.IpAddress).String(), tcpaddr.Port)
	}
	return ""
}

func (cc *ChannelzClient) ListServerSockets(ctx context.Context) {
	now := timeNow()

	cc.printf("%s\t%s\t%-40s\t%-20s\t%-20s\t%-20s\t%s\t%s\t%s\t%s\n",
		"ID", "ServerID", "Name", "RemoteName", "Local", "Remote", "Started", "Success", "Fail", "LastStream")

	cc.visitGetServers(ctx, func(server *channelzpb.Server) {
		cc.visitGetServerSockets(ctx, server.Ref.ServerId, func(socket *channelzpb.Socket) {
			localIP := addrToString(socket.Local)
			remoteIP := addrToString(socket.Remote)

			cc.printf("%d\t%-8d\t%-40s\t%-20s\t%-16s\t%-16s\t%-6d\t%-6d\t%-6d\t%-8s\n",
				socket.Ref.SocketId,
				server.Ref.ServerId,
				decorateEmpty(socket.Ref.Name),
				decorateEmpty(socket.RemoteName),
				decorateEmpty(localIP),
				decorateEmpty(remoteIP),
				socket.Data.StreamsStarted,
				socket.Data.StreamsSucceeded,
				socket.Data.StreamsFailed,
				elapsedTimestamp(now, socket.Data.LastRemoteStreamCreatedTimestamp),
			)
		})
	})
}

func (cc *ChannelzClient) visitGetServerSockets(ctx context.Context, id int64, fn func(*channelzpb.Socket)) {
	lastSocketID := int64(0)
	for {
		res, err := cc.cc.GetServerSockets(ctx, &channelzpb.GetServerSocketsRequest{
			ServerId:      id,
			StartSocketId: lastSocketID,
		})
		if err != nil {
			log.Fatalf("err: %v\n", err)
		}

		for _, ref := range res.SocketRef {
			socket, err := cc.cc.GetSocket(ctx, &channelzpb.GetSocketRequest{SocketId: ref.SocketId})
			if err != nil {
				log.Fatalf("err %v\n", err)
			}

			fn(socket.Socket)
		}
		if res.End {
			break
		}

		lastSocketID++
	}
}

func (cc *ChannelzClient) TreeTopChannels(ctx context.Context) {
	now := timeNow()

	cc.visitTopChannels(ctx, func(channel *channelzpb.Channel) {
		cc.printf("%s (ID:%d) [%s]\n",
			channel.Data.Target, channel.Ref.ChannelId,
			channel.Data.State.State.String())

		elapesed := elapsedTimestamp(now, channel.Data.LastCallStartedTimestamp)
		cc.printf("  [Calls] Started:%v, Succeeded:%v, Failed:%v, Last:%v\n", channel.Data.CallsStarted, channel.Data.CallsSucceeded, channel.Data.CallsFailed, elapesed)

		// for _, ev := range channel.Data.Trace.Events {
		// 	cc.printf("ev %v\n", ev)
		// }

		for _, socket := range channel.SocketRef {
			cc.printf("socket %v\n", socket)
		}

		for _, ch := range channel.ChannelRef {
			cc.printf("ch %v\n", ch)
		}

		if len(channel.SubchannelRef) != 0 {
			cc.printf("  [Subchannels]\n")
		}
		for _, ch := range channel.SubchannelRef {
			res, err := cc.cc.GetSubchannel(ctx, &channelzpb.GetSubchannelRequest{SubchannelId: ch.SubchannelId})
			if err != nil {
				log.Fatalf("err %v", err)
			}

			subch := res.Subchannel
			cc.printf("    |-- %s (ID:%d) [%s]\n",
				subch.Data.Target, subch.Ref.SubchannelId,
				subch.Data.State.State.String())

			elapesed := elapsedTimestamp(now, subch.Data.LastCallStartedTimestamp)
			cc.printf("          [Calls]: Started:%v, Succeeded:%v, Failed:%v, Last:%s\n", subch.Data.CallsStarted, subch.Data.CallsSucceeded, subch.Data.CallsFailed, elapesed)

			for _, socket := range subch.SocketRef {
				res, err := cc.cc.GetSocket(ctx, &channelzpb.GetSocketRequest{SocketId: socket.SocketId})
				if err != nil {
					log.Fatalf("err %v\n", err)
				}

				socket := res.Socket
				cc.printf("          [Socket] ID:%v, Name:%v, RemoteName:%v", socket.Ref.SocketId, socket.Ref.Name, socket.RemoteName)
				cc.printf(", Local:")
				if addr := socket.Local.GetTcpipAddress(); addr != nil {
					cc.printf("[%v]:%v", net.IP(addr.IpAddress).String(), addr.Port)
				}
				cc.printf(" Remote:")
				if addr := socket.Remote.GetTcpipAddress(); addr != nil {
					cc.printf("[%v]:%v", net.IP(addr.IpAddress).String(), addr.Port)
				}
				cc.printf("\n")
			}

			for _, ch := range subch.ChannelRef {
				cc.printf("---- ch %v\n", ch)
			}
			for _, ch := range subch.SubchannelRef {
				cc.printf("---- ch %v\n", ch)
			}
		}

		cc.printf("\n")
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
			if found == nil {
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

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second*time.Duration(retry+1))
		defer cancel()
		res, err := cc.cc.GetTopChannels(ctx, &channelzpb.GetTopChannelsRequest{StartChannelId: lastChannelID})
		if err != nil {
			retry++
			continue
		}

		for _, channel := range res.Channel {
			fn(channel)
			if id := channel.GetRef().GetChannelId(); id > lastChannelID {
				lastChannelID = id
			}
		}
		if res.End {
			break
		}

		lastChannelID++
	}
}
