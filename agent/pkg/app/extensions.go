package app

import (
	"sort"

	"github.com/up9inc/mizu/agent/pkg/controllers"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"

	amqpExt "github.com/up9inc/mizu/tap/extensions/amqp"
	httpExt "github.com/up9inc/mizu/tap/extensions/http"
	kafkaExt "github.com/up9inc/mizu/tap/extensions/kafka"
	redisExt "github.com/up9inc/mizu/tap/extensions/redis"
)

var (
	Extensions    []*tapApi.Extension          // global
	ExtensionsMap map[string]*tapApi.Extension // global
)

func LoadExtensions() {
	Extensions = make([]*tapApi.Extension, 4)
	ExtensionsMap = make(map[string]*tapApi.Extension)

	extensionAmqp := &tapApi.Extension{}
	dissectorAmqp := amqpExt.NewDissector()
	dissectorAmqp.Register(extensionAmqp)
	extensionAmqp.Dissector = dissectorAmqp
	Extensions[0] = extensionAmqp
	ExtensionsMap[extensionAmqp.Protocol.Name] = extensionAmqp

	extensionHttp := &tapApi.Extension{}
	dissectorHttp := httpExt.NewDissector()
	dissectorHttp.Register(extensionHttp)
	extensionHttp.Dissector = dissectorHttp
	Extensions[1] = extensionHttp
	ExtensionsMap[extensionHttp.Protocol.Name] = extensionHttp

	extensionKafka := &tapApi.Extension{}
	dissectorKafka := kafkaExt.NewDissector()
	dissectorKafka.Register(extensionKafka)
	extensionKafka.Dissector = dissectorKafka
	Extensions[2] = extensionKafka
	ExtensionsMap[extensionKafka.Protocol.Name] = extensionKafka

	extensionRedis := &tapApi.Extension{}
	dissectorRedis := redisExt.NewDissector()
	dissectorRedis.Register(extensionRedis)
	extensionRedis.Dissector = dissectorRedis
	Extensions[3] = extensionRedis
	ExtensionsMap[extensionRedis.Protocol.Name] = extensionRedis

	sort.Slice(Extensions, func(i, j int) bool {
		return Extensions[i].Protocol.Priority < Extensions[j].Protocol.Priority
	})

	for _, extension := range Extensions {
		logger.Log.Infof("Extension Properties: %+v", extension)
	}

	controllers.InitExtensionsMap(ExtensionsMap)
}
