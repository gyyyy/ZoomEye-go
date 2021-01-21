package main

import (
	"ZoomEye-go/zoomeye"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type config struct {
	ConfigPath string `yaml:"ZOOMEYE_CONFIG_PATH"`
	CachePath  string `yaml:"ZOOMEYE_CACHE_PATH"`
	DataPath   string `yaml:"ZOOMEYE_DATA_PATH"`
	ExpiredSec uint   `yaml:"EXPIRED_TIME"`
}

func (c *config) check() {
	if c.ConfigPath == "" {
		c.ConfigPath = "~/.config/zoomeye/setting"
	}
	if c.CachePath == "" {
		c.CachePath = "~/.config/zoomeye/cache"
	}
	if c.DataPath == "" {
		c.DataPath = "data"
	}
	if c.ExpiredSec == 0 {
		c.ExpiredSec = 432000
	}
	if checkFolder(&c.ConfigPath) == nil {
		if b, err := yaml.Marshal(c); err == nil {
			writeFile(filepath.Join(c.ConfigPath, "conf.yaml"), b)
		}
	}
	checkFolder(&c.CachePath)
	checkFolder(&c.DataPath)
}

func newConfig() *config {
	conf := &config{}
	defer conf.check()
	var (
		path   = "conf.yaml"
		b, err = readFile(path)
	)
	if err != nil {
		if !os.IsNotExist(err) {
			return conf
		}
		path = filepath.Join("~/.config/zoomeye/setting", path)
		if b, err = readFile(path); err != nil && !os.IsNotExist(err) {
			return conf
		}
	}
	if b != nil && len(b) > 0 {
		yaml.Unmarshal(b, conf)
	}
	return conf
}

func hash(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// ZoomEyeAgent represents agent of ZoomEye
type ZoomEyeAgent struct {
	zoom *zoomeye.ZoomEye
	conf *config
}

// InitByKey initializes ZoomEye by API-Key
func (a *ZoomEyeAgent) InitByKey(apiKey string) (*zoomeye.ResourcesInfoResult, error) {
	var (
		zoom        = zoomeye.NewWithKey(apiKey, "")
		result, err = zoom.ResourcesInfo()
	)
	if err != nil {
		return nil, err
	}
	if err = writeFile(filepath.Join(a.conf.ConfigPath, "apikey"), []byte(apiKey)); err != nil {
		return nil, err
	}
	a.zoom = zoom
	return result, nil
}

// InitByUser initializes ZoomEye by username/password
func (a *ZoomEyeAgent) InitByUser(username, password string) (*zoomeye.ResourcesInfoResult, error) {
	var (
		zoom     = zoomeye.New()
		tok, err = zoom.Login(username, password)
	)
	if err != nil {
		return nil, err
	}
	result, err := zoom.ResourcesInfo()
	if err != nil {
		return nil, err
	}
	if err = writeFile(filepath.Join(a.conf.ConfigPath, "jwt"), []byte(tok)); err != nil {
		return nil, err
	}
	a.zoom = zoom
	return result, nil
}

func (a *ZoomEyeAgent) loadAuthKey(name string) (string, error) {
	var (
		path      = filepath.Join(a.conf.ConfigPath, name)
		info, err = os.Stat(path)
	)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(fmt.Sprintf("%o", info.Mode()), "600") {
		os.Chmod(path, 0o600)
	}
	b, err := readFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// InitLocal initializes ZoomEye from local Key files
func (a *ZoomEyeAgent) InitLocal() (*zoomeye.ResourcesInfoResult, error) {
	var (
		result              *zoomeye.ResourcesInfoResult
		apiKey, accessToken string
		err                 error
	)
	if apiKey, err = a.loadAuthKey("apikey"); err != nil {
		if accessToken, err = a.loadAuthKey("jwt"); err != nil {
			return nil, fmt.Errorf("cannot find any valid Auth Key file")
		}
	}
	zoom := zoomeye.NewWithKey(apiKey, accessToken)
	if result, err = zoom.ResourcesInfo(); err != nil {
		return nil, err
	}
	a.zoom = zoom
	return result, nil
}

// Info gets resources information
func (a *ZoomEyeAgent) Info() (*zoomeye.ResourcesInfoResult, error) {
	if a.zoom == nil {
		return a.InitLocal()
	}
	return a.zoom.ResourcesInfo()
}

func (a *ZoomEyeAgent) isExpiredData(t time.Time) bool {
	return time.Now().Sub(t) > (time.Duration(a.conf.ExpiredSec) * time.Second)
}

func (a *ZoomEyeAgent) fromLocal(name string) (*zoomeye.SearchResult, bool) {
	path := filepath.Join(a.conf.DataPath, name)
	if info, err := os.Stat(path); err != nil || a.isExpiredData(info.ModTime()) {
		return nil, false
	}
	b, err := readFile(path)
	if err != nil {
		return nil, false
	}
	result := &zoomeye.SearchResult{}
	err = json.Unmarshal(b, result)
	return result, err == nil
}

func (a *ZoomEyeAgent) hasCached(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if a.isExpiredData(info.ModTime()) {
		os.Remove(path)
		return false
	}
	return true
}

func (a *ZoomEyeAgent) fromCache(name string) (*zoomeye.SearchResult, bool) {
	path := filepath.Join(a.conf.CachePath, name)
	if !a.hasCached(path) {
		return nil, false
	}
	b, err := readFile(path)
	if err != nil {
		return nil, false
	}
	result := &zoomeye.SearchResult{}
	if err = json.Unmarshal(b, result); err != nil {
		return nil, false
	}
	return result, true
}

func (a *ZoomEyeAgent) cache(name string, result *zoomeye.SearchResult) error {
	var (
		path   = filepath.Join(a.conf.CachePath, name)
		b, err = json.Marshal(result)
	)
	if err != nil {
		return err
	}
	return writeFile(path, b)
}

// Search gets search results from local, cache or API
func (a *ZoomEyeAgent) Search(dork string, num int, resource string, force bool) (*zoomeye.SearchResult, error) {
	if a.zoom == nil {
		if _, err := a.InitLocal(); err != nil {
			return nil, err
		}
	}
	if num <= 0 {
		num = 20
	}
	if resource != "" {
		resource = strings.ToLower(resource)
	} else {
		resource = "host"
	}
	maxPage := num / 20
	if num%20 > 0 {
		maxPage++
	}
	if force {
		result, err := a.zoom.MultiPageSearch(dork, maxPage, resource, "")
		if err != nil {
			return nil, err
		}
		if num < len(result.Matches) {
			result.Matches = result.Matches[:num]
		}
		return result, nil
	}
	result, ok := a.fromLocal(fmt.Sprintf("%s_%s_%d", resource, url.QueryEscape(dork), num))
	if ok {
		result.Type = resource
		return result, nil
	}
	result = &zoomeye.SearchResult{
		Type: resource,
	}
	for i := 0; i < maxPage; i++ {
		var (
			res  = &zoomeye.SearchResult{}
			page = i + 1
			name = hash(fmt.Sprintf("%s_%s_%d", resource, dork, page)) + ".json"
		)
		if res, ok = a.fromCache(name); !ok {
			var err error
			if res, err = a.zoom.DorkSearch(dork, page, resource, ""); err != nil {
				return nil, err
			}
			if err = a.cache(name, res); err != nil {
				return nil, err
			}
		}
		result.Extend(res)
	}
	if num < len(result.Matches) {
		result.Matches = result.Matches[:num]
	}
	return result, nil
}

// Save writes the search results (and filter data) to local file
func (a *ZoomEyeAgent) Save(result *zoomeye.SearchResult, name string) (string, error) {
	var (
		path      = filepath.Join(a.conf.DataPath, name+".json")
		data, err = json.Marshal(result)
	)
	if err != nil {
		data = []byte("{}")
	}
	if err = writeFile(path, data); err != nil {
		return "", err
	}
	if len(result.FilterCache) > 0 {
		if data, err = json.Marshal(result.FilterCache); err == nil {
			writeFile(filepath.Join(a.conf.DataPath, name+"_filtered.json"), data)
		}
	}
	path, _ = filepath.Abs(path)
	return path, nil
}

// NewAgent creates instance of ZoomEyeAgent
func NewAgent() *ZoomEyeAgent {
	return &ZoomEyeAgent{
		conf: newConfig(),
	}
}
