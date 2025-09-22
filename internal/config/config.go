package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type TmuxWindow struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
}

type Config struct {
	WorktreesLocation string       `toml:"worktrees_location"`
	CopyFiles         []string     `toml:"copy_files"`
	TmuxWindows       []TmuxWindow `toml:"tmux_windows"`
}

var defaultConfig = Config{
	WorktreesLocation: filepath.Join(os.Getenv("HOME"), "projects", "worktrees"),
	CopyFiles:         []string{},
	TmuxWindows:       []TmuxWindow{},
}

var currentConfig *Config

func Load() (*Config, error) {
	if currentConfig != nil {
		return currentConfig, nil
	}

	config := defaultConfig

	// Load global config
	var globalConfig Config
	globalConfigPath := getGlobalConfigPath()
	if err := loadConfigFile(globalConfigPath, &globalConfig); err == nil {
		// Merge global config
		if globalConfig.WorktreesLocation != "" {
			config.WorktreesLocation = globalConfig.WorktreesLocation
		}
		config.CopyFiles = append(config.CopyFiles, globalConfig.CopyFiles...)
		config.TmuxWindows = append(config.TmuxWindows, globalConfig.TmuxWindows...)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading global config: %w", err)
	}

	// Load local config
	var localConfig Config
	localConfigPath := getLocalConfigPath()
	if err := loadConfigFile(localConfigPath, &localConfig); err == nil {
		// Local config overrides
		if localConfig.WorktreesLocation != "" {
			config.WorktreesLocation = localConfig.WorktreesLocation
		}
		// Merge copy_files arrays (local adds to global)
		config.CopyFiles = append(config.CopyFiles, localConfig.CopyFiles...)
		// Merge tmux_windows arrays (local adds to global)
		config.TmuxWindows = append(config.TmuxWindows, localConfig.TmuxWindows...)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading local config: %w", err)
	}

	// Remove duplicates from CopyFiles
	config.CopyFiles = removeDuplicates(config.CopyFiles)

	config.WorktreesLocation = expandPath(config.WorktreesLocation)

	currentConfig = &config
	return currentConfig, nil
}

func removeDuplicates(files []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, file := range files {
		if !seen[file] {
			seen[file] = true
			result = append(result, file)
		}
	}
	return result
}

func loadConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if _, err := toml.Decode(string(data), config); err != nil {
		return fmt.Errorf("error parsing config file %s: %w", path, err)
	}

	return nil
}

func getGlobalConfigPath() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(xdgConfigHome, "wt", "config.toml")
}

func getLocalConfigPath() string {
	return ".wt.toml"
}

func expandPath(path string) string {
	if path == "" {
		return path
	}

	if path[:2] == "~/" {
		home := os.Getenv("HOME")
		path = filepath.Join(home, path[2:])
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	return absPath
}

func GetWorktreesLocation() string {
	config, err := Load()
	if err != nil {
		return defaultConfig.WorktreesLocation
	}
	return config.WorktreesLocation
}

func GetCopyFiles() []string {
	config, err := Load()
	if err != nil {
		return defaultConfig.CopyFiles
	}
	return config.CopyFiles
}

func GetTmuxWindows() []TmuxWindow {
	config, err := Load()
	if err != nil {
		return defaultConfig.TmuxWindows
	}
	return config.TmuxWindows
}

func CreateGlobalConfigDir() error {
	configPath := getGlobalConfigPath()
	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}
