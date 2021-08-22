package configStructs

const (
	GuiPortViewName = "gui-port"
)

type ViewConfig struct {
	GuiPort uint16 `yaml:"gui-port" default:"8899"`
}
