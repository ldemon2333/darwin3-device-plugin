package deviceplugin

import (
	"context"
	"path"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/ldemon/Darwin3-device-plugin/pkg/common"
	"github.com/pkg/errors"
)

func (d *Darwin3DevicePlugin) Register() error {
	conn, err := connect(pluginapi.KubeletSocket, common.ConnectTimeout)
	if err != nil {
		return errors.WithMessagef(err, "failed to connect to kubelet socket %s", pluginapi.KubeletSocket)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(common.DeviceSocket),
		ResourceName: common.ResourceName,
	}

	if _, err := client.Register(context.Background(), req); err != nil {
		return errors.WithMessagef(err, "failed to register device plugin %s", common.ResourceName)
	}

	return nil
}
