# channelzcli

channelzcli is a command line tool for gRPC channelz service.

## Commands

### List

`list` command displays a table of information about the specified type.

Avaiable types:

* `channel`: shows root channels
* `server`: shows servers in the proccess


```
$ channelzcli -k --addr localhost:8000 list channel
ID	Name                                    	Channel	SubChannel	Calls	Success	Fail	LastCall
1	spanner.googleapis.com:443              	0      	1         	3444  	3436  	8     	1m      
2	spanner.googleapis.com:443              	0      	1         	3451  	3444  	9     	34s     
3	spanner.googleapis.com:443              	0      	1         	3315  	3306  	11    	8s      
4	spanner.googleapis.com:443              	0      	1         	3724  	3714  	13    	2m      
28	pubsub.googleapis.com:443               	0      	16        	0     	0     	0     	none    
29	pubsub.googleapis.com:443               	0      	16        	40    	40    	0     	13h     
```

```
$ channelzcli -k --addr localhost:8000 list server
ID	Name	LocalAddr	Calls	Success	Fail	LastCall
31	<none>	<none>      	2264  	2262  	1     	410ms
35	<none>	[::]:5000   	1732  	1090  	642   	10h
```

### Describe

`describe` command displays details about the specified type.

Avaiable types:

* `channel`
* `server`


```
$ channelzcli -k --addr localhost:8000 describe channel spanner.googleapis.com:443
$ channelzcli -k --addr localhost:8000 describe channel 4
```

```
$ channelzcli -k --addr localhost:8000 describe server foo
$ channelzcli -k --addr localhost:8000 describe server 31
```

### Tree


`tree` command displays a tree of information about the specified type recursively.

Available types:

* `channel`
* `server`


```
$ channelzcli -k --addr localhost:8000 tree channel
pubsub.googleapis.com:443 (ID:28) [READY]
  [Calls] Started:0, Succeeded:0, Failed:0, Last:none
  [Subchannels]
    |-- pubsub.googleapis.com:443 (ID:40) [READY]
          [Calls]: Started:0, Succeeded:0, Failed:0, Last:none
          [Socket] ID:11562, Name:, RemoteName:, Local:[10.0.0.2]:47708 Remote:[172.217.26.42]:443
    |-- pubsub.googleapis.com:443 (ID:46) [READY]
          [Calls]: Started:0, Succeeded:0, Failed:0, Last:none
          [Socket] ID:11557, Name:, RemoteName:, Local:[10.0.0.2]:34138 Remote:[172.217.161.74]:443
    |-- pubsub.googleapis.com:443 (ID:41) [READY]
          [Calls]: Started:0, Succeeded:0, Failed:0, Last:none
          [Socket] ID:11552, Name:, RemoteName:, Local:[10.0.0.2]:60344 Remote:[216.58.197.138]:443
    |-- pubsub.googleapis.com:443 (ID:52) [READY]
          [Calls]: Started:0, Succeeded:0, Failed:0, Last:none
          [Socket] ID:11561, Name:, RemoteName:, Local:[10.0.0.2]:47706 Remote:[172.217.26.42]:443
    |-- pubsub.googleapis.com:443 (ID:43) [READY]
          [Calls]: Started:0, Succeeded:0, Failed:0, Last:none
          [Socket] ID:11556, Name:, RemoteName:, Local:[10.0.0.2]:34142 Remote:[172.217.161.74]:443
```

## How to run channelz server (in Go)

* Use [RegisterChannelzServiceToServer](https://godoc.org/google.golang.org/grpc/channelz/service#RegisterChannelzServiceToServer) to register channelz service to gRPC server
* Require grpc-go v1.15.0 or later
* It's also usefull for gRPC client only application, not serving gRPC server, to expose client metrics


```go
import (
	"log"
	"net"

	"google.golang.org/grpc"
	channelzsvc "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

func main() {
	s := grpc.NewServer()
	reflection.Register(s)
	channelzsvc.RegisterChannelzServiceToServer(s)

	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := s.Serve(lis); err != nil {
		log.Fatalf("err %v\n", err)
	}
}
```
