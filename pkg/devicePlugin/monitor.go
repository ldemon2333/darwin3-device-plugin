package deviceplugin

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type DeviceMonitor struct {
	path    string
	devices map[string]*pluginapi.Device
	notify  chan struct{} //notify when device update
}

func NewDeviceMonitor(path string) *DeviceMonitor {
	return &DeviceMonitor{
		path:    path,
		devices: make(map[string]*pluginapi.Device),
		notify:  make(chan struct{}),
	}
}

func (dm *DeviceMonitor) List() error {
	err := filepath.Walk(dm.path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			klog.Infof("Skipping directory: %s", path)
			return nil
		}
		dm.devices[info.Name()] = &pluginapi.Device{
			ID:     info.Name(),
			Health: pluginapi.Healthy,
		}
		return nil
	})
	return errors.WithMessagef(err, "failed to list devices in path %s", dm.path)
}

func (dm *DeviceMonitor) Watch() error {
	// This is a placeholder for the watch logic.
	// In a real implementation, this would monitor the device path for changes
	// and update the devices map accordingly, notifying via dm.notify channel.
	klog.Infof("Watching for device changes in path: %s", dm.path)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithMessagef(err, "failed to create fsnotify watcher for path %s", dm.path)
	}
	defer w.Close()

	errChan := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- errors.Errorf("panic in fsnotify watcher: %v", r)
			}
		}()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					continue
				}
				klog.Infof("fsnotify event: %s %s", event.Op.String(), event.Name)

				if event.Op&fsnotify.Create == fsnotify.Create {
					dev := path.Base(event.Name)
					dm.devices[dev] = &pluginapi.Device{
						ID:     dev,
						Health: pluginapi.Healthy,
					}
					dm.notify <- struct{}{}
					klog.Infof("find new device [%s]", dev)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					dev := path.Base(event.Name)
					delete(dm.devices, dev)
					dm.notify <- struct{}{}
					klog.Infof("device [%s] removed", dev)
				}
			case err, ok := <-w.Errors:
				if !ok {
					continue
				}
				klog.Errorf("fsnotify watch device failed: %v", err)
			}

		}
	}()
	err = w.Add(dm.path)
	if err != nil {
		return errors.WithMessagef(err, "failed to add path %s to fsnotify watcher", dm.path)
	}
	return <-errChan
}

func (dm *DeviceMonitor) Devices() []*pluginapi.Device {
	devices := make([]*pluginapi.Device, 0, len(dm.devices))
	for _, device := range dm.devices {
		devices = append(devices, device)
	}
	return devices
}

func String(devs []*pluginapi.Device) string {
	ids := make([]string, 0, len(devs))
	for _, dev := range devs {
		ids = append(ids, dev.ID)
	}
	return strings.Join(ids, ",")

}
