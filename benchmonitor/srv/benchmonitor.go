package main

import (
	"flag"
	"fmt"
	//"io"
	"log"
	"net"
	"os/exec"

	"google.golang.org/grpc"

	pb "github.com/kkourt/kubenetbench/benchmonitor/api"
)

var (
	srvPort = flag.Int("p", 8451, "Server port")
)

type monitorSrv struct {
	pb.UnimplementedKubebenchMonitorServer
}

func (*monitorSrv) GetSysInfo(
	_ *pb.Empty,
	stream pb.KubebenchMonitor_GetSysInfoServer,
) error {

	cmd := exec.Command("scripts/system_info.sh")
	/*
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StdoutPipe()
		cmd.Start()
		buff := make([]byte, 1024)
		for {
			n, err := stdout.Read(buff)
			if err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("io error: %w", err)
			}

			err = stream.Send(&pb.File{
				Data: buff[:n],
			})

			if err != nil {
				return fmt.Errorf("io error: %w", err)
			}
		}
	*/
	data, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("io error: %w", err)
	}
	err = stream.Send(&pb.File{
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("io error: %w", err)
	}

	return nil
}

func newMonitorSrv() *monitorSrv {
	return &monitorSrv{}
}

func main() {
	log.Println("starting monitor server")
	flag.Parse()

	laddr := fmt.Sprintf(":%d", *srvPort)
	listen, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(fmt.Errorf("listen (%s) failed: %w", laddr, err))
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterKubebenchMonitorServer(grpcSrv, newMonitorSrv())
	grpcSrv.Serve(listen)
}
