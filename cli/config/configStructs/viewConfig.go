package configStructs

const (
	GuiPortViewName   = "gui-port"
	UrlViewName       = "url"
)

type ViewConfig struct {
	GuiPort   uint16 `yaml:"gui-port" default:"8899"`
	Url       string `yaml:"url,omitempty" readonly:""`
}
