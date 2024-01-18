# About

Sample grpc server and client.

Communication transport: unix domain socket.

The goal of this sample is to implement monitoring of socket file deletion and if socket file is deleted - restart grpc server and recreate socket file.

# Build

```
go build -o srv ./server/main.go
```

```
go build -o clnt ./client/main.go
```

# Run

In separate terminal run server:
```
./srv
```

In separate terminal run client:
```
./clnt
```

Try to remove socket file `rm /tmp/echo.sock` and then run client again - server should be restarted and communicating and socket file present again.

# Links

Used code samples from:  

https://grpc.io/docs/languages/go/quickstart/

https://github.com/jianfengye/go-superviser