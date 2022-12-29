package kubernetes

type Resources struct {
	CpuLimit       string `yaml:"cpu-limit" default:"750m"`
	MemoryLimit    string `yaml:"memory-limit" default:"1Gi"`
	CpuRequests    string `yaml:"cpu-requests" default:"50m"`
	MemoryRequests string `yaml:"memory-requests" default:"50Mi"`
}
