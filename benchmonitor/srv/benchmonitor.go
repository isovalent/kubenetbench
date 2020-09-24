package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"

	"google.golang.org/grpc"

	pb "github.com/cilium/kubenetbench/benchmonitor/api"
)

var (
	srvPort = flag.Int("p", 8451, "Server port")
)

type monitorSrv struct {
	pb.UnimplementedKubebenchMonitorServer
	pendingCmds sync.Map
}

type ErrCmdInProgress struct{}

func (e *ErrCmdInProgress) Error() string {
	return fmt.Sprintf("command still running")
}

func (srv *monitorSrv) StartCollection(
	ctx context.Context,
	arg *pb.CollectionConf,
) (*pb.Empty, error) {

	ret := &pb.Empty{}
	cid := arg.CollectionId
	_, loaded := srv.pendingCmds.LoadOrStore(cid, &ErrCmdInProgress{})

	if loaded {
		return ret, fmt.Errorf(fmt.Sprintf("id %s already exists", cid))
	}

	go func() {
		t := arg.Duration
		cmd := exec.Command("/scripts/perf-record.sh", t, cid)
		err := cmd.Run()
		srv.pendingCmds.Store(cid, err)
	}()

	return ret, nil
}

type FileSender interface {
	Send(*pb.File) error
}

func copyFileToStream(fname string, stream FileSender) error {

	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	buff := make([]byte, 4096)
	for {
		n, err := f.Read(buff)
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

	return nil
}

func (srv *monitorSrv) GetCollectionResults(
	arg *pb.CollectionResultsConf,
	stream pb.KubebenchMonitor_GetCollectionResultsServer,
) error {
	cid := arg.CollectionId
	cmd_err, ok := srv.pendingCmds.Load(cid)
	if !ok {
		return fmt.Errorf(fmt.Sprintf("invalid collection id %s", cid))
	}

	switch cmd_err.(type) {
	case ErrCmdInProgress:
		return fmt.Errorf("command still running: %w", cmd_err)
	}

	srv.pendingCmds.Delete(cid)
	if cmd_err != nil {
		return fmt.Errorf("command resulted in error: %w", cmd_err)
	}

	cmd := exec.Command("/scripts/perf-collect.sh", cid)
	collect_err := cmd.Run()
	if collect_err != nil {
		return fmt.Errorf("collection (%s) command resulted in error: %w", cmd, collect_err)
	}

	fname := fmt.Sprintf("/tmp/%s-perf.data.tar.bz2", cid)
	return copyFileToStream(fname, stream)
}

func (*monitorSrv) GetSysInfo(
	_ *pb.Empty,
	stream pb.KubebenchMonitor_GetSysInfoServer,
) error {

	cmd := exec.Command("scripts/system_info.sh")
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
