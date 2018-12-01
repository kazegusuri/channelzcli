package channelz

import (
	"context"
	"fmt"
	"log"
	"net"
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

func (cc *ChannelzClient) GetServers(ctx context.Context) {
	lastServerID := int64(0)
	now := time.Now()
	for {
		res, err := cc.cc.GetServers(ctx, &channelzpb.GetServersRequest{StartServerId: lastServerID})
		if err != nil {
			log.Fatalf("err: %v\n", err)
		}

		for _, server := range res.Server {
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

			if lastServerID < server.Ref.ServerId {
				lastServerID = server.Ref.ServerId
			}

			fmt.Println("")
		}
		if res.End {
			break
		}

		lastServerID++
	}
}

func (cc *ChannelzClient) GetTopChannels(ctx context.Context) {
	lastChannelID := int64(0)
	now := time.Now()
	for {
		res, err := cc.cc.GetTopChannels(ctx, &channelzpb.GetTopChannelsRequest{StartChannelId: lastChannelID})
		if err != nil {
			log.Fatalf("err: %v\n", err)
		}

		for _, channel := range res.Channel {
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

			if lastChannelID < channel.Ref.ChannelId {
				lastChannelID = channel.Ref.ChannelId
			}

			fmt.Println()
		}
		if res.End {
			break
		}

		lastChannelID++
	}

}
