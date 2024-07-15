package netplan

import (
	"fmt"
	"os"
	"path/filepath"
)

type InterfaceType int

const InterfaceTypeUnknown InterfaceType = -1

const (
	InterfaceTypeLoopback InterfaceType = iota
	InterfaceTypeEthernet
	InterfaceTypeBonding
	InterfaceTypeBridge
	InterfaceTypeVirtual
	InterfaceTypeVLAN
)

var InterfaceTypeToString = map[InterfaceType]string{
	InterfaceTypeUnknown:  "Unknown",
	InterfaceTypeLoopback: "Loopback",
	InterfaceTypeEthernet: "Ethernet",
	InterfaceTypeBonding:  "Bonding",
	InterfaceTypeBridge:   "Bridge",
	InterfaceTypeVirtual:  "Virtual",
	InterfaceTypeVLAN:     "VLAN",
}

func (n InterfaceType) String() string {
	v, ok := InterfaceTypeToString[n]
	if !ok {
		return InterfaceTypeToString[InterfaceTypeUnknown]
	}

	return v
}

func getDriver(interfaceName string) (string, error) {
	driverPath := fmt.Sprintf("/sys/class/net/%s/device/driver", interfaceName)

	// No driver (possibly virtual or special interface)
	if _, err := os.Lstat(driverPath); os.IsNotExist(err) {
		return "", nil
	}

	driverLink, err := filepath.EvalSymlinks(driverPath)
	if err != nil {
		return "", err
	}
	driverName := filepath.Base(driverLink)
	return driverName, nil
}

func getInterfaceType(interfaceName string) (InterfaceType, error) {
	basePath := fmt.Sprintf("/sys/class/net/%s", interfaceName)

	driver, err := getDriver(interfaceName)
	if driver != "" && err == nil {
		return InterfaceTypeEthernet, nil
	}

	if _, err := os.Lstat(filepath.Join(basePath, "bridge")); !os.IsNotExist(err) {
		return InterfaceTypeBridge, nil
	}

	if _, err := os.Lstat(filepath.Join(basePath, "bonding")); !os.IsNotExist(err) {
		return InterfaceTypeBonding, nil
	}

	if _, err := os.ReadFile(filepath.Join(basePath, "master")); err == nil {
		return InterfaceTypeVLAN, nil
	}

	if interfaceName == "lo" {
		return InterfaceTypeLoopback, nil
	}

	virtualPath := fmt.Sprintf("/sys/devices/virtual/net/%s", interfaceName)
	if _, err := os.Lstat(virtualPath); !os.IsNotExist(err) {
		return InterfaceTypeVirtual, nil
	}

	return InterfaceTypeUnknown, nil
}
