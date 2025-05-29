package main

import (
	deviceplugin "github.com/ldemon/Darwin3-device-plugin/pkg/devicePlugin"
	"github.com/ldemon/Darwin3-device-plugin/pkg/utils"
	"k8s.io/klog/v2"
)

func main() {
	klog.Infof("Starting Darwin3 DevicePlugin...")
	dp := deviceplugin.NewDarwin3DevicePlugin()

	go dp.Run()

	if err := dp.Register(); err != nil {
		klog.Fatalf("Failed to register device plugin: %v", err)
	}

	stop := make(chan struct{})
	err := utils.WatchKubelet(stop)
	if err != nil {
		klog.Fatalf("Failed to watch kubelet: %v", err)
	}
	<-stop
	klog.Infof("kubelet restart, exiting...")
}
