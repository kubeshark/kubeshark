package configStructs

const (
	GuiPortViewName   = "gui-port"
	UrlViewName       = "url"
	ProxyTypeViewName = "proxy-type"
)

type ViewConfig struct {
	GuiPort   uint16 `yaml:"gui-port" default:"8899"`
	ProxyType string `yaml:"proxy-type" default:"proxy"`
	Url       string `yaml:"url,omitempty" readonly:""`
}
