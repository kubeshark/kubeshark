package elastic

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap/api"
	"net/http"
	"sync"
	"time"
)

type client struct {
	es            *elasticsearch.Client
	index         string
	insertedCount int
}

var instance *client
var once sync.Once

func GetInstance() *client {
	once.Do(func() {
		instance = newClient()
	})
	return instance
}

func (client *client) Configure(config shared.ElasticConfig) {
	if config.Url == "" || config.User == "" || config.Password == "" {
		logger.Log.Infof("No elastic configuration was supplied, elastic exporter disabled")
		return
	}
	transport := http.DefaultTransport
	tlsClientConfig := &tls.Config{InsecureSkipVerify: true}
	transport.(*http.Transport).TLSClientConfig = tlsClientConfig
	cfg := elasticsearch.Config{
		Addresses: []string{config.Url},
		Username:  config.User,
		Password:  config.Password,
		Transport: transport,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logger.Log.Fatalf("Failed to initialize elastic client %v", err)
	}

	// Have the client instance return a response
	res, err := es.Info()
	if err != nil {
		logger.Log.Fatalf("Elastic client.Info() ERROR: %v", err)
	} else {
		client.es = es
		client.index = "mizu_traffic_http_" + time.Now().Format("2006_01_02_15_04")
		client.insertedCount = 0
		logger.Log.Infof("Elastic client configured, index: %s, cluster info: %v", client.index, res)
	}
	defer res.Body.Close()
}

func newClient() *client {
	return &client{
		es:    nil,
		index: "",
	}
}

type httpEntry struct {
	Source      *api.TCP               `json:"src"`
	Destination *api.TCP               `json:"dst"`
	Outgoing    bool                   `json:"outgoing"`
	CreatedAt   time.Time              `json:"createdAt"`
	Request     map[string]interface{} `json:"request"`
	Response    map[string]interface{} `json:"response"`
	Summary     string                 `json:"summary"`
	Method      string                 `json:"method"`
	Status      int                    `json:"status"`
	ElapsedTime int64                  `json:"elapsedTime"`
	Path        string                 `json:"path"`
}

func (client *client) PushEntry(entry *api.Entry) {
	if client.es == nil {
		return
	}

	if entry.Protocol.Name != "http" {
		return
	}

	entryToPush := httpEntry{
		Source:      entry.Source,
		Destination: entry.Destination,
		Outgoing:    entry.Outgoing,
		CreatedAt:   entry.StartTime,
		Request:     entry.Request,
		Response:    entry.Response,
		Summary:     entry.Summary,
		Method:      entry.Method,
		Status:      entry.Status,
		ElapsedTime: entry.ElapsedTime,
		Path:        entry.Path,
	}

	entryJson, err := json.Marshal(entryToPush)
	if err != nil {
		logger.Log.Errorf("json.Marshal ERROR: %v", err)
		return
	}
	var buffer bytes.Buffer
	buffer.WriteString(string(entryJson))
	res, _ := client.es.Index(client.index, &buffer)
	if res.StatusCode == 201 {
		client.insertedCount += 1
	}
}
