package tap

type globalSettings struct {
	filterPorts       []int
	filterAuthorities []string
}

var gSettings = &globalSettings{
	filterPorts:       []int{},
	filterAuthorities: []string{},
}

func SetFilterPorts(ports []int) {
	gSettings.filterPorts = ports
}

func GetFilterPorts() []int {
	ports := make([]int, len(gSettings.filterPorts))
	copy(ports, gSettings.filterPorts)
	return ports
}

func SetFilterAuthorities(ipAddresses []string) {
	gSettings.filterAuthorities = ipAddresses
}

func GetFilterIPs() []string {
	addresses := make([]string, len(gSettings.filterAuthorities))
	copy(addresses, gSettings.filterAuthorities)
	return addresses
}
