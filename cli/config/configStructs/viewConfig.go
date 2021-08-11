package configStructs

const (
	GuiPortViewName        = "gui-port"
	KubeConfigPathViewName = "kube-config"
)

type ViewConfig struct {
	GuiPort        uint16 `yaml:"gui-port" default:"8899"`
	KubeConfigPath string `yaml:"kube-config"`
}
