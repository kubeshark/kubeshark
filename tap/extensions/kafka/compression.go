package kafka

import (
	"github.com/segmentio/kafka-go/compress"
)

type Compression = compress.Compression

type CompressionCodec = compress.Codec

// TODO: this file should probably go away once the internals of the package
// have moved to use the protocol package.
