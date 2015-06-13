package git4go

import (
	"errors"
	"fmt"
	"github.com/Unknwon/goconfig"
	"os"
	"path/filepath"
	"strings"
)

type ConfigLevel int

const (
	GIT_CONFIG_FILENAME_SYSTEM string = "gitconfig"
	GIT_CONFIG_FILENAME_GLOBAL string = ".gitconfig"
	GIT_CONFIG_FILENAME_XDG    string = "config"
	GIT_CONFIG_FILENAME_INREPO string = "config"

	ConfigLevelSystem  ConfigLevel = 1
	ConfigLevelXDG     ConfigLevel = 2
	ConfigLevelGlobal  ConfigLevel = 3
	ConfigLevelLocal   ConfigLevel = 4
	ConfigLevelApp     ConfigLevel = 5
	ConfigLevelHighest ConfigLevel = -1
)

/*type ConfigEntry struct {
	Name  string
	Value string
	Level ConfigLevel
}*/

// Repository method related to Config

func (repo *Repository) Config() *Config {
	if repo.config == nil {
		config, _ := NewConfig()
		path := filepath.Join(repo.pathRepository, GIT_CONFIG_FILENAME_INREPO)
		_, err := os.Stat(path)
		if os.IsExist(err) {
			err = config.AddFile(path, ConfigLevelLocal, false)
			if err != nil {
				return nil
			}
		}
		path, err = ConfigFindGlobal()
		if err == nil {
			err = config.AddFile(path, ConfigLevelGlobal, false)
			if err != nil {
				return nil
			}
		}
		path, err = ConfigFindXDG()
		if err == nil {
			err = config.AddFile(path, ConfigLevelXDG, false)
			if err != nil {
				return nil
			}
		}
		path, err = ConfigFindSystem()
		if err == nil {
			err = config.AddFile(path, ConfigLevelSystem, false)
			if err != nil {
				return nil
			}
		}
		repo.config = config
	}
	return repo.config
}

// Config type and its methods

type configFile struct {
	force bool
	level ConfigLevel
	file  *goconfig.ConfigFile
}

type Config struct {
	files []*configFile
}

func NewConfig() (*Config, error) {
	config := new(Config)
	return config, nil
}

func (c *Config) AddFile(path string, level ConfigLevel, force bool) error {
	file, err := goconfig.LoadConfigFile(path)
	if err != nil {
		return err
	}
	entry := &configFile{
		force: force,
		level: level,
		file:  file,
	}
	c.files = append(c.files, entry)
	return nil
}

func (c *Config) LookupInt32(name string) (int32, error) {
	keys := strings.SplitN(name, ".", 2)
	for _, file := range c.files {
		value, err := file.file.Int(keys[0], keys[1])
		if err == nil {
			return int32(value), nil
		}
	}
	return 0, errors.New(fmt.Sprintf("Config value '%s' was not found", name))
}

func (c *Config) LookupInt64(name string) (int64, error) {
	keys := strings.SplitN(name, ".", 2)
	for _, file := range c.files {
		value, err := file.file.Int64(keys[0], keys[1])
		if err == nil {
			return value, nil
		}
	}
	return 0, errors.New(fmt.Sprintf("Config value '%s' was not found", name))
}

func (c *Config) LookupString(name string) (string, error) {
	keys := strings.SplitN(name, ".", 2)
	for _, file := range c.files {
		value, err := file.file.GetValue(keys[0], keys[1])
		if err == nil {
			return value, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Config value '%s' was not found", name))
}

func (c *Config) LookupBool(name string) (bool, error) {
	keys := strings.SplitN(name, ".", 2)
	for _, file := range c.files {
		value, err := file.file.Bool(keys[0], keys[1])
		if err == nil {
			return value, nil
		}
	}
	return false, errors.New(fmt.Sprintf("Config value '%s' was not found", name))
}

func (c *Config) SetString(name, value string) (err error) {
	return nil
}

func (c *Config) SetInt32(name string, value int32) (err error) {
	return nil
}

func (c *Config) SetInt64(name string, value int64) (err error) {
	return nil
}

func (c *Config) SetBool(name string, value bool) (err error) {
	return nil
}

func ConfigFindGlobal() (string, error) {
	return findInDirList(GIT_CONFIG_FILENAME_GLOBAL, "global")
}

func ConfigFindSystem() (string, error) {
	return findInDirList(GIT_CONFIG_FILENAME_SYSTEM, "system")
}

func ConfigFindXDG() (string, error) {
	return findInDirList(GIT_CONFIG_FILENAME_XDG, "global/xdg")
}
