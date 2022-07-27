package dependency

type ContainerType string

const (
	ServiceMapGeneratorDependency ContainerType = "ServiceMapGeneratorDependency"
	OasGeneratorDependency        ContainerType = "OasGeneratorDependency"
	EntriesInserter               ContainerType = "EntriesInserter"
	EntriesProvider               ContainerType = "EntriesProvider"
	EntriesSocketStreamer         ContainerType = "EntriesSocketStreamer"
	EntryStreamerSocketConnector  ContainerType = "EntryStreamerSocketConnector"
)
