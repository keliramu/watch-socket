package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	pb "github.com/keliramu/watch-socket/helloworld"

	"google.golang.org/grpc"
)

var (
	end         = make(chan bool)
	sockPath    = "/tmp/echo.sock"
	monitorPath = "/tmp"
)

type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	// Cleanup the sockfile.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Exiting by sgnal...")
		os.Remove(sockPath)
		os.Exit(1)
	}()

	// channel for delete signal
	r := make(chan int)
	go func() {
		for {
			s, l := createServer()
			go startServer(s, l)
			select {
			case <-end:
				return
			case <-r:
				s.Stop()
			}
		}
	}()

	// watch for changes
	monitor(monitorPath, r)
}

func startServer(s *grpc.Server, l net.Listener) {
	if err := s.Serve(l); err != nil {
		log.Fatalln(err)
	}
}

func createServer() (*grpc.Server, net.Listener) {
	// Create a Unix domain socket and listen for incoming connections.
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	return s, l
}

func monitor(root string, dc chan int) {
	fd, err := syscall.InotifyInit()
	if fd == -1 || err != nil {
		end <- true
		return
	}

	flags := syscall.IN_MODIFY | syscall.IN_CREATE | syscall.IN_DELETE // 2/128/512
	wd, _ := syscall.InotifyAddWatch(fd, root, uint32(flags))
	if wd == -1 {
		end <- true
		return
	}
	var (
		buf [syscall.SizeofInotifyEvent * 10]byte
		n   int
	)

	for {
		n, _ = syscall.Read(fd, buf[0:])
		if n > syscall.SizeofInotifyEvent {
			var offset = 0
			for offset < n {
				raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
				mask := uint32(raw.Mask)
				offset = offset + int(raw.Len) + syscall.SizeofInotifyEvent

				switch mask {
				case syscall.IN_DELETE:
					if _, err := os.Stat(sockPath); err != nil {
						log.Println("action: DEL:", sockPath)
						dc <- 1
					}
				}
			}
		}
	}
}
