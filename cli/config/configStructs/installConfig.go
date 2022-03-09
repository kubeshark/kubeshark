package configStructs

type InstallConfig struct {
	TemplateUrl  string `yaml:"template-url" default:"https://storage.googleapis.com/static.up9.io/mizu/helm-template"`
	TemplateName string `yaml:"template-name" default:"helm-template.yaml"`
}
