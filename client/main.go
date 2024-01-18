// Package main implements a client for Greeter service.
package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/keliramu/watch-socket/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	name     = "tobesent"
	protocol = "unix"
	sockPath = "/tmp/echo.sock"
)

var (
	credentials = insecure.NewCredentials() // No SSL/TLS
	dialer      = func(ctx context.Context, addr string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, protocol, addr)
	}
	options = []grpc.DialOption{
		grpc.WithTransportCredentials(credentials),
		grpc.WithBlock(),
		grpc.WithContextDialer(dialer),
	}
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(sockPath, options...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
