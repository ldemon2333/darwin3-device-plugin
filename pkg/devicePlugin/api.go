package deviceplugin

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func (d *Darwin3DevicePlugin) GetDevicePluginOptions(_ context.Context, _ *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	// Return empty options as we don't have any specific options to provide
	return &pluginapi.DevicePluginOptions{}, nil
}

func (d *Darwin3DevicePlugin) GetPreferredAllocation(ctx context.Context, req *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	klog.Infoln("[GetPreferredAllocation] running with request:", req)
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (d *Darwin3DevicePlugin) ListAndWatch(_ *pluginapi.Empty, stream pluginapi.DevicePlugin_ListAndWatchServer) error {
	// This method is used to stream the list of devices to the kubelet
	// In a real implementation, you would send updates when devices change
	devs := d.dm.Devices()
	klog.Infof("find devices [%s]", String(devs))

	err := stream.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	if err != nil {
		return errors.WithMessage(err, "failed to send ListAndWatch response")
	}

	klog.Infof("ListAndWatch response sent successfully, waiting for device updates...")
	for range d.dm.notify {
		devs = d.dm.Devices()
		klog.Infof("Device update detected, sending updated devices: [%s]", String(devs))
		_ = stream.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	}
	return nil
}

func (d *Darwin3DevicePlugin) Allocate(_ context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	ret := &pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		klog.Infof("[Allocate] running with request: %v", strings.Join(req.DevicesIDs, ","))
		resp := &pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"Darwin3": strings.Join(req.DevicesIDs, ","),
			},
		}
		ret.ContainerResponses = append(ret.ContainerResponses, resp)
	}
	return ret, nil
}

func (d *Darwin3DevicePlugin) PreStartContainer(_ context.Context, _ *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	// This method is called before starting a container that uses the device plugin
	// In this example, we do not have any specific pre-start actions
	klog.Infoln("[PreStartContainer] No pre-start actions defined")
	return &pluginapi.PreStartContainerResponse{}, nil
}
