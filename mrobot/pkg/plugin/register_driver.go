package plugin

import (
	"fmt"
	"github.com/zibuyu28/cmapp/mrobot/drivers"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RegisterDriver register driver build in
func RegisterDriver(d drivers.BuildInDriver) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading RPC server: %s\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	grpcserver := grpc.NewServer()

	driver.RegisterMachineDriverServer(grpcserver, d.GrpcServer)

	go grpcserver.Serve(listener)

	fmt.Println(listener.Addr())

	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			time.Sleep(time.Second)
			// driver exit
			d.Exit()
			return
		case syscall.SIGHUP:
		// TODO app reload
		default:
			return
		}
	}
}
