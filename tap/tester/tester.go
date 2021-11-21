package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"plugin"
	"sort"
	"strings"

	"github.com/op/go-logging"

	"github.com/go-errors/errors"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/tap"
	tapApi "github.com/up9inc/mizu/tap/api"
)

func loadExtensions() ([]*tapApi.Extension, error) {
	extensionsDir := "./extensions"
	files, err := ioutil.ReadDir(extensionsDir)

	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	extensions := make([]*tapApi.Extension, 0)
	for _, file := range files {
		filename := file.Name()

		if !strings.HasSuffix(filename, ".so") {
			continue
		}

		logger.Log.Infof("Loading extension: %s", filename)

		extension := &tapApi.Extension{
			Path: path.Join(extensionsDir, filename),
		}

		plug, err := plugin.Open(extension.Path)

		if err != nil {
			return nil, errors.Wrap(err, 0)
		}

		extension.Plug = plug
		symDissector, err := plug.Lookup("Dissector")

		if err != nil {
			return nil, errors.Wrap(err, 0)
		}

		dissector, ok := symDissector.(tapApi.Dissector)

		if !ok {
			return nil, errors.Errorf("Symbol Dissector type error: %v %T", file, symDissector)
		}

		dissector.Register(extension)
		extension.Dissector = dissector
		extensions = append(extensions, extension)
	}

	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Protocol.Priority < extensions[j].Protocol.Priority
	})

	for _, extension := range extensions {
		logger.Log.Infof("Extension Properties: %+v", extension)
	}

	return extensions, nil
}

func internalRun() error {
	logger.InitLoggerStderrOnly(logging.DEBUG)

	opts := tap.TapOpts{
		HostMode: false,
	}

	outputItems := make(chan *tapApi.OutputChannelItem, 1000)
	extenssions, err := loadExtensions()

	if err != nil {
		return err
	}

	tapOpts := tapApi.TrafficFilteringOptions{}

	tap.StartPassiveTapper(&opts, outputItems, extenssions, &tapOpts)

	logger.Log.Infof("Tapping, press enter to exit...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadLine()
	return nil
}

func main() {
	err := internalRun()

	if err != nil {
		switch err := err.(type) {
		case *errors.Error:
			logger.Log.Errorf("Error: %v", err.ErrorStack())
		default:
			logger.Log.Errorf("Error: %v", err)
		}

		os.Exit(1)
	}
}
