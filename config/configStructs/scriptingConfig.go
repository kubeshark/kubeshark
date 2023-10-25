package configStructs

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog/log"
)

type ScriptingConfig struct {
	Env          map[string]interface{} `yaml:"env" json:"env" default:"{}"`
	Source       string                 `yaml:"source" json:"source" default:""`
	WatchScripts bool                   `yaml:"watchScripts" json:"watchScripts" default:"true"`
}

func (config *ScriptingConfig) GetScripts() (scripts []*misc.Script, err error) {
	if config.Source == "" {
		return
	}

	var files []fs.DirEntry
	files, err = os.ReadDir(config.Source)
	if err != nil {
		return
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		var script *misc.Script
		path := filepath.Join(config.Source, f.Name())
		script, err = misc.ReadScriptFile(path)
		if err != nil {
			return
		}
		scripts = append(scripts, script)

		log.Debug().Str("path", path).Msg("Found script:")
	}

	return
}
