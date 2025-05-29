package common

import "time"

const (
	ResourceName   string = "zj.com/darwin3"
	DevicePath     string = "/etc/darwin3"
	DeviceSocket   string = "darwin3-device-plugin.sock"
	ConnectTimeout        = 5 * time.Second
)
