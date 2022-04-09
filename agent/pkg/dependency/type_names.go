package dependency

type DependencyContainerType string

const (
	ServiceMapGeneratorDependency = "ServiceMapGeneratorDependency"
	OasGeneratorDependency        = "OasGeneratorDependency"
	EntriesProvider               = "EntriesProvider"
	EntriesSocketStreamer         = "EntriesSocketStreamer"
	EntryStreamerSocketConnector  = "EntryStreamerSocketConnector"
)
