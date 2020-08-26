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

func (s *Session) GetSysInfoNode(node string) error {
	srvAddr := fmt.Sprintf("%s:%s", node, "8451")
	conn, err := grpc.Dial(srvAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to monitor %s: %w", srvAddr, err)
	}
	defer conn.Close()

	fname := fmt.Sprintf("%s/%s.sysinfo", s.dir, node)
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer f.Close()

	cli := pb.NewKubebenchMonitorClient(conn)
	stream, err := cli.GetSysInfo(context.Background(), &pb.Empty{})

	if err != nil {
		return fmt.Errorf("failed to retrieve sysinfo from monitor %s: %w", srvAddr, err)
	}

	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to retrieve sysinfo from monitor %s: %w", srvAddr, err)
		}

		_, err = f.Write(data.Data)
		if err != nil {
			return fmt.Errorf("Error writing data: %w", err)
		}
	}

	return err
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
