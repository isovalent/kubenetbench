package core

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"text/template"
	"time"

	"google.golang.org/grpc"

	pb "github.com/kkourt/kubenetbench/benchmonitor/api"
)

var monitorTemplate = template.Must(template.New("monitor").Parse(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: knb-monitor
  labels:
    {{.sessLabel}}
    role: monitor
spec:
  selector:
    matchLabels:
      {{.sessLabel}}
      role: monitor
  template:
    metadata:
      labels:
        {{.sessLabel}}
        role: monitor
    spec:
      tolerations:
      # this toleration is to have the daemonset runnable on master nodes
      # remove it if your masters can't run pods
      - key: node-role.kubernetes.io/master
        effect: NoSchedule

      #
      hostNetwork: true
      hostPID: true
      hostIPC: true

      containers:
      - name: kubenetbench-monitor
        image: docker.io/kkourt/kubenetbench-monitor
        securityContext:
           privileged: true
           capabilities:
              add:
                 # - NET_ADMIN
                 - SYS_ADMIN
        ports:
           - containerPort: 8451
             hostPort: 8451
        volumeMounts:
        - name: tmp
          mountPath: /tmp
      volumes:
      - name: tmp
        hostPath:
          path: /tmp
`))

func (s *Session) genMonitorYaml() (string, error) {
	yaml := fmt.Sprintf("%s/monitor.yaml", s.dir)
	log.Printf("Generating %s", yaml)
	f, err := os.Create(yaml)
	if err != nil {
		return "", err
	}

	vals := map[string]interface{}{
		"sessLabel": s.getSessionLabel(": "),
	}
	err = monitorTemplate.Execute(f, vals)
	if err != nil {
		return "", err
	}
	f.Close()
	return yaml, nil
}

type FileReceiver interface {
	Recv() (*pb.File, error)
}

func copyStreamToFile(fname string, stream FileReceiver) error {

	f, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("io error: %w", err)
		}

		_, err = f.Write(data.Data)
		if err != nil {
			return fmt.Errorf("Error writing data: %w", err)
		}
	}

	return nil
}

func (s *Session) GetSysInfoNode(node string) error {
	srvAddr := fmt.Sprintf("%s:%s", node, "8451")
	conn, err := grpc.Dial(srvAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to monitor %s: %w", srvAddr, err)
	}
	defer conn.Close()

	cli := pb.NewKubebenchMonitorClient(conn)
	stream, err := cli.GetSysInfo(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to retrieve sysinfo from monitor %s: %w", srvAddr, err)
	}

	fname := fmt.Sprintf("%s/%s.sysinfo", s.dir, node)
	return copyStreamToFile(fname, stream)
}

func (s *Session) GetSysInfoNodes() error {

	lines, err := KubeGetNodes()
	if err != nil {
		return err
	}

	retriesOrig := 10
	for _, node := range lines {
		retries := retriesOrig
		for {
			log.Printf("calling GetSysInfoNode on %s (remaining retries: %d)", node, retries)
			err = s.GetSysInfoNode(node)
			if err == nil {
				break
			}

			if retries == 0 {
				return fmt.Errorf("Error calling GetSysInfoNode %s after %d retries (last error:%w)", node, retriesOrig, err)
			}

			retries--
			time.Sleep(4 * time.Second)
		}
	}

	return nil
}

func (r *RunBenchCtx) endCollection() error {
	var err error = nil

	for _, node := range r.collectNodes {
		srvAddr := fmt.Sprintf("%s:%s", node, "8451")
		conn, err := grpc.Dial(srvAddr, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("failed to connect to monitor %s: %w", srvAddr, err)
		}
		defer conn.Close()
		cli := pb.NewKubebenchMonitorClient(conn)
		conf := &pb.CollectionResultsConf{
			CollectionId: r.runid,
		}

		stream, err := cli.GetCollectionResults(context.Background(), conf)
		if err != nil {
			log.Printf("collection on monitor %s failed: %s\n", node, err)
		}

		fname := fmt.Sprintf("%s/perf-%s.tar.bz2", r.getDir(), node)
		err = copyStreamToFile(fname, stream)
		if err != nil {
			log.Printf("writing collection data from node %s failed: %s\n", node, err)
		} else {
			log.Printf("perf data for %s can be found in: %s\n", node, fname)
		}
	}

	return err
}

func (r *RunBenchCtx) startCollection() error {

	labels := [...]string{PodName, PodNodeName, PodPhase}
	podsinfo, err := r.KubeGetPods__(labels[:])
	if err != nil {
		return err
	}

	nodes := make(map[string]struct{})
	log.Printf("Pods: \n")
	for _, a := range podsinfo {
		log.Printf(" %v\n", a)
		nodes[a[1]] = struct{}{}
	}

	for node, _ := range nodes {
		srvAddr := fmt.Sprintf("%s:%s", node, "8451")
		conn, err := grpc.Dial(srvAddr, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("failed to connect to monitor %s: %w", srvAddr, err)
		}
		defer conn.Close()
		//log.Printf("connected to monitor on %s\n", node)
		cli := pb.NewKubebenchMonitorClient(conn)
		conf := &pb.CollectionConf{
			Duration:     "5",
			CollectionId: r.runid,
		}

		_, err = cli.StartCollection(context.Background(), conf)
		if err == nil {
			log.Printf("started collection on monitor %s\n", node)
			r.collectNodes = append(r.collectNodes, node)
		} else {
			log.Printf("started collection on monitor %s failed: %s\n", node, err)
		}
	}

	return nil
}
