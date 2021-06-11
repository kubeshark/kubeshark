package tap

type globalSettings struct {
	filterPorts       []int
	filterIpAddresses []string
}

var gSettings = &globalSettings{
	filterPorts:       []int{},
	filterIpAddresses: []string{},
}

func SetFilterPorts(ports []int) {
	gSettings.filterPorts = ports
}

func GetFilterPorts() []int {
	ports := make([]int, len(gSettings.filterPorts))
	copy(ports, gSettings.filterPorts)
	return ports
}

func SetFilterIPs(ipAddresses []string) {
	gSettings.filterIpAddresses = ipAddresses
}

func GetFilterIPs() []string {
	addresses := make([]string, len(gSettings.filterIpAddresses))
	copy(addresses, gSettings.filterIpAddresses)
	return addresses
}
