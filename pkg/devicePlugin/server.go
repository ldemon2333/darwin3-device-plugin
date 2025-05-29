package deviceplugin

import (
	"context"
	"log"
	"net"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/ldemon/Darwin3-device-plugin/pkg/common"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type Darwin3DevicePlugin struct {
	server *grpc.Server
	stop   chan struct{} // this channel signals to stop the device plugin
	dm     *DeviceMonitor
}

func NewDarwin3DevicePlugin() *Darwin3DevicePlugin {
	return &Darwin3DevicePlugin{
		server: grpc.NewServer(grpc.EmptyServerOption{}),
		stop:   make(chan struct{}),
		dm:     NewDeviceMonitor(common.DevicePath),
	}
}

func (c *Darwin3DevicePlugin) Run() error {
	err := c.dm.List()
	if err != nil {
		log.Fatalf("Failed to list devices: %v", err)
	}

	go func() {
		if err := c.dm.Watch(); err != nil {
			log.Println("watch error:", err)
		}
	}()

	pluginapi.RegisterDevicePluginServer(c.server, c)

	socket := path.Join(pluginapi.DevicePluginPath, common.DeviceSocket)
	err = syscall.Unlink(socket)
	if err != nil && !os.IsNotExist(err) {
		return errors.WithMessagef(err, "failed to unlink socket %s", socket)
	}

	sock, err := net.Listen("unix", socket)
	if err != nil {
		return errors.WithMessagef(err, "failed to listen on socket %s", socket)
	}

	go c.server.Serve(sock)

	conn, err := connect(common.DeviceSocket, 5*time.Second)
	if err != nil {
		return errors.WithMessagef(err, "failed to connect to device plugin socket %s", common.DeviceSocket)
	}
	conn.Close()
	return nil
}

func connect(socket string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			if deadline, ok := ctx.Deadline(); ok {
				return net.DialTimeout("unix", addr, time.Until(deadline))
			}
			return net.DialTimeout("unix", addr, common.ConnectTimeout)
		}),
	}

	conn, err := grpc.DialContext(ctx, socket, opts...)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to connect to socket %s", socket)
	}

	return conn, nil
}
