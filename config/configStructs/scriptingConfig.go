package configStructs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog/log"
)

type ScriptingConfig struct {
	Env          map[string]interface{} `yaml:"env" json:"env" default:"{}"`
	Source       string                 `yaml:"source" json:"source" default:""`
	Sources      []string               `yaml:"sources" json:"sources" default:"[]"`
	WatchScripts bool                   `yaml:"watchScripts" json:"watchScripts" default:"true"`
	Active       []string               `yaml:"active" json:"active" default:"[]"`
	Console      bool                   `yaml:"console" json:"console" default:"true"`
}

func (config *ScriptingConfig) GetScripts() (scripts []*misc.Script, err error) {
	// Check if both Source and Sources are empty
	if config.Source == "" && len(config.Sources) == 0 {
		return nil, nil
	}

	var allFiles []struct {
		Source string
		File   fs.DirEntry
	}

	// Handle single Source directory
	if config.Source != "" {
		files, err := os.ReadDir(config.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %v", config.Source, err)
		}
		for _, file := range files {
			allFiles = append(allFiles, struct {
				Source string
				File   fs.DirEntry
			}{Source: config.Source, File: file})
		}
	}

	// Handle multiple Sources directories
	if len(config.Sources) > 0 {
		for _, source := range config.Sources {
			files, err := os.ReadDir(source)
			if err != nil {
				return nil, fmt.Errorf("failed to read directory %s: %v", source, err)
			}
			for _, file := range files {
				allFiles = append(allFiles, struct {
					Source string
					File   fs.DirEntry
				}{Source: source, File: file})
			}
		}
	}

	// Iterate over all collected files
	for _, f := range allFiles {
		if f.File.IsDir() {
			continue
		}

		// Construct the full path based on the relevant source directory
		path := filepath.Join(f.Source, f.File.Name())
		if !strings.HasSuffix(f.File.Name(), ".js") { // Use file name suffix for skipping non-JS files
			log.Info().Str("path", path).Msg("Skipping non-JS file")
			continue
		}

		// Read the script file
		var script *misc.Script
		script, err = misc.ReadScriptFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read script file %s: %v", path, err)
		}

		// Append the valid script to the scripts slice
		scripts = append(scripts, script)

		log.Debug().Str("path", path).Msg("Found script:")
	}

	// Return the collected scripts and nil error if successful
	return scripts, nil
}
