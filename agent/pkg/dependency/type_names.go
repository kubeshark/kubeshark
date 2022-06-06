package dependency

type DependencyContainerType string

const (
	ServiceMapGeneratorDependency = "ServiceMapGeneratorDependency"
	OasGeneratorDependency        = "OasGeneratorDependency"
	EntriesInserter               = "EntriesInserter"
	EntriesProvider               = "EntriesProvider"
	EntriesSocketStreamer         = "EntriesSocketStreamer"
	EntryStreamerSocketConnector  = "EntryStreamerSocketConnector"
)
