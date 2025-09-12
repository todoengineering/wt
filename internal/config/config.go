package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	WorktreesLocation string `toml:"worktrees_location"`
}

var defaultConfig = Config{
	WorktreesLocation: filepath.Join(os.Getenv("HOME"), "projects", "worktrees"),
}

var currentConfig *Config

func Load() (*Config, error) {
	if currentConfig != nil {
		return currentConfig, nil
	}

	config := defaultConfig

	globalConfigPath := getGlobalConfigPath()
	if err := loadConfigFile(globalConfigPath, &config); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading global config: %w", err)
	}

	localConfigPath := getLocalConfigPath()
	if err := loadConfigFile(localConfigPath, &config); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading local config: %w", err)
	}

	config.WorktreesLocation = expandPath(config.WorktreesLocation)

	currentConfig = &config
	return currentConfig, nil
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

func CreateGlobalConfigDir() error {
	configPath := getGlobalConfigPath()
	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}