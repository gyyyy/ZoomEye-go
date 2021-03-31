package main

import (
	"crypto/md5"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gyyyy/ZoomEye-go/zoomeye"
	"gopkg.in/yaml.v2"
)

// NoAuthKeyErr represents error of no any Auth Keys
type NoAuthKeyErr struct {
	msg string
}

func (e *NoAuthKeyErr) Error() string {
	return e.msg
}

func noAuthKey(err error) *NoAuthKeyErr {
	return &NoAuthKeyErr{
		msg: err.Error(),
	}
}

type config struct {
	ConfigPath string `yaml:"ZOOMEYE_CONFIG_PATH"`
	CachePath  string `yaml:"ZOOMEYE_CACHE_PATH"`
	DataPath   string `yaml:"ZOOMEYE_DATA_PATH"`
	ExpiredSec uint   `yaml:"EXPIRED_TIME"`
}

func (c *config) check() {
	if c.ConfigPath == "" {
		c.ConfigPath = abs("~/.config/zoomeye/setting")
	}
	if c.CachePath == "" {
		c.CachePath = abs("~/.config/zoomeye/cache")
	}
	if c.DataPath == "" {
		c.DataPath = "data"
	}
	if c.ExpiredSec == 0 {
		c.ExpiredSec = 432000
	}
	if checkFolder(c.ConfigPath) == nil && checkFolder(c.CachePath) == nil && checkFolder(c.DataPath) == nil {
		if b, err := yaml.Marshal(c); err == nil {
			writeFile(filepath.Join(c.ConfigPath, "conf.yml"), b)
		}
	}
}

func newConfig() *config {
	conf := &config{
		ConfigPath: abs("~/.config/zoomeye/setting"),
	}
	defer conf.check()
	var (
		path   = "conf.yml"
		b, err = readFile(path)
	)
	if err != nil {
		if !os.IsNotExist(err) {
			return conf
		}
		if b, err = readFile(filepath.Join(conf.ConfigPath, path)); err != nil && !os.IsNotExist(err) {
			return conf
		}
	}
	if b != nil && len(b) > 0 {
		yaml.Unmarshal(b, conf)
	}
	return conf
}

func filename(resource, dork string, n int, enc bool) string {
	name := fmt.Sprintf("%s_%s_%d", resource, dork, n)
	if enc {
		name = fmt.Sprintf("%x", md5.Sum([]byte(name)))
	}
	return name + ".json"
}

// ZoomEyeAgent represents agent of ZoomEye
type ZoomEyeAgent struct {
	zoom *zoomeye.ZoomEye
	conf *config
}

func (a *ZoomEyeAgent) isExpiredData(t time.Time) bool {
	return time.Now().Sub(t) > (time.Duration(a.conf.ExpiredSec) * time.Second)
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

func (a *ZoomEyeAgent) fromCache(name string, result interface{}) bool {
	path := filepath.Join(a.conf.CachePath, name)
	return a.hasCached(path) && readObject(result, path) == nil
}

func (a *ZoomEyeAgent) cache(name string, result interface{}) error {
	return writeObject(filepath.Join(a.conf.CachePath, name), result)
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
		return "", noAuthKey(err)
	}
	if !strings.HasSuffix(fmt.Sprintf("%o", info.Mode()), "600") {
		os.Chmod(path, 0o600)
	}
	b, err := readFile(path)
	if err != nil {
		return "", noAuthKey(err)
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
			return nil, err
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

func (a *ZoomEyeAgent) fromLocal(name string) (*zoomeye.SearchResult, bool) {
	path := filepath.Join(a.conf.DataPath, name)
	if info, err := os.Stat(path); err != nil || a.isExpiredData(info.ModTime()) {
		return nil, false
	}
	result := &zoomeye.SearchResult{}
	if readObject(result, path) != nil {
		return nil, false
	}
	return result, true
}

func (a *ZoomEyeAgent) forceSearch(dork string, maxPage int, resource string) (*zoomeye.SearchResult, error) {
	results, err := a.zoom.MultiPageSearch(dork, maxPage, resource, "")
	if err != nil {
		return nil, err
	}
	result := &zoomeye.SearchResult{
		Type: resource,
	}
	for i, n := 0, len(results); i < maxPage && n > 0; i++ {
		page := i + 1
		if res, ok := results[page]; ok {
			a.cache(filename(resource, dork, page, true), res)
			result.Extend(res)
			n--
		}
	}
	return result, nil
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
	maxPage := num / 20
	if num%20 > 0 {
		maxPage++
	}
	if resource = strings.ToLower(resource); resource != "web" {
		resource = "host"
	}
	if force {
		return a.forceSearch(dork, maxPage, resource)
	}
	result, ok := a.fromLocal(filename(resource, url.QueryEscape(dork), num, false))
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
			name = filename(resource, dork, page, true)
		)
		if !a.fromCache(name, res) {
			var err error
			if res, err = a.zoom.DorkSearch(dork, page, resource, ""); err != nil {
				return nil, err
			}
			a.cache(name, res)
		}
		result.Extend(res)
	}
	if num < len(result.Matches) {
		result.Matches = result.Matches[:num]
	}
	return result, nil
}

// Load reads local data, and unmarshals to search results
func (a *ZoomEyeAgent) Load(path string) (*zoomeye.SearchResult, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	result := &zoomeye.SearchResult{
		Type: "host",
	}
	if err := readObject(result, path); err != nil {
		return nil, err
	}
	if len(result.Matches) > 0 && result.Matches[0].Find("site") != nil {
		result.Type = "web"
	}
	return result, nil
}

// History gets query results of device history by IP
func (a *ZoomEyeAgent) History(ip string, force bool) (*zoomeye.HistoryResult, error) {
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid ip address")
	}
	info, err := a.Info()
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(info.Plan) {
	case "user", "developer":
		return nil, fmt.Errorf("this function is only open to advanced users and VIP users.")
	}
	var (
		result *zoomeye.HistoryResult
		name   = fmt.Sprintf("%x", md5.Sum([]byte("history_"+ip))) + ".json"
		ok     bool
	)
	if !force {
		result = &zoomeye.HistoryResult{}
		ok = a.fromCache(name, result)
	}
	if !ok {
		if result, err = a.zoom.HistoryIP(ip); err != nil {
			return nil, err
		}
		if result.Count == 0 || len(result.Data) == 0 {
			return result, nil
		}
		a.cache(name, result)
	}
	for i := 0; i < len(result.Data); {
		if _, ok := result.Data[i]["component"]; ok {
			result.Data = append(result.Data[:i], result.Data[i+1:]...)
		} else {
			i++
		}
	}
	result.Count = uint64(len(result.Data))
	return result, nil
}

// Clear removes all cache or setting data
func (a *ZoomEyeAgent) Clear(cache, setting bool) {
	if cache {
		os.RemoveAll(a.conf.CachePath)
	}
	if setting {
		os.RemoveAll(a.conf.ConfigPath)
		os.MkdirAll(a.conf.ConfigPath, os.ModePerm)
		if b, err := yaml.Marshal(a.conf); err == nil {
			writeFile(filepath.Join(a.conf.ConfigPath, "conf.yml"), b)
		}
	}
}

// SaveFilterData writes the filter data to local file
func (a *ZoomEyeAgent) SaveFilterData(path string, data []map[string]interface{}) error {
	if data == nil || len(data) == 0 {
		return fmt.Errorf("no any filter datas")
	}
	return writeObject(path, data)
}

// Save writes the search results (and filter data) to local file
func (a *ZoomEyeAgent) Save(name string, result *zoomeye.SearchResult) (string, error) {
	path := filepath.Join(a.conf.DataPath, name+".json")
	if err := writeObject(path, result); err != nil {
		return "", err
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
