package git4go

import (
	"errors"
	"fmt"
	"github.com/Unknwon/goconfig"
	"os"
	"path/filepath"
	"strings"
	"strconv"
)

type ConfigLevel int

const (
	ConfigFileNameSystem string = "gitconfig"
	ConfigFileNameGlobal string = ".gitconfig"
	ConfigFileNameXDG    string = "config"
	ConfigFileNameInrepo string = "config"

	ConfigLevelSystem  ConfigLevel = 1
	ConfigLevelXDG     ConfigLevel = 2
	ConfigLevelGlobal  ConfigLevel = 3
	ConfigLevelLocal   ConfigLevel = 4
	ConfigLevelApp     ConfigLevel = 5
	ConfigLevelHighest ConfigLevel = -1
)

// todo: implement cache

/*type ConfigEntry struct {
	Name  string
	Value string
	Level ConfigLevel
}*/

// Repository method related to Config

func (repo *Repository) Config() *Config {
	if repo.config == nil {
		config, _ := NewConfig()
		path := filepath.Join(repo.pathRepository, ConfigFileNameInrepo)
		_, err := os.Stat(path)
		if !os.IsNotExist(err) {
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

func (c *Config) LookupStringWithDefaultValue(name string) (string, error) {
	result, err := c.LookupString(name)
	if err == nil {
		return result, nil
	}
	result, ok := defaultStringConfig[name]
	if ok {
		return result, nil
	}
	return "", err
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

func (c *Config) LookupBooleanWithDefaultValue(name string) (bool, error) {
	result, err := c.LookupBool(name)
	if err == nil {
		return result, nil
	}
	result, ok := defaultBoolConfig[name]
	if ok {
		return result, nil
	}
	return false, err
}

func (c *Config) SetString(name, value string) (err error) {
	if len(c.files) > 0 && c.files[0].level == ConfigLevelLocal {
		file := c.files[0].file
		keys := strings.SplitN(name, ".", 2)
		file.SetValue(keys[0], keys[1], value)
		path, err := ConfigFindGlobal()
		if err != nil {
			return err
		}
		goconfig.SaveConfigFile(file, path)
	}
	return nil
}

func (c *Config) SetInt32(name string, value int32) (err error) {
	return c.SetString(name, strconv.Itoa(int(value)))
}

func (c *Config) SetInt64(name string, value int64) (err error) {
	return c.SetString(name, strconv.Itoa(int(value)))
}

func (c *Config) SetBool(name string, value bool) (err error) {
	if value {
		return c.SetString(name, "true")
	} else {
		return c.SetString(name, "false")
	}
}

func ConfigFindGlobal() (string, error) {
	return findInDirList(ConfigFileNameGlobal, "global")
}

func ConfigFindSystem() (string, error) {
	return findInDirList(ConfigFileNameSystem, "system")
}

func ConfigFindXDG() (string, error) {
	return findInDirList(ConfigFileNameXDG, "global/xdg")
}
