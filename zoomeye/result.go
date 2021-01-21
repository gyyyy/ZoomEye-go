package zoomeye

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var (
	filterFields = map[string]map[string]string{
		"host": map[string]string{
			"_index":     "ip",
			"app":        "portinfo.app",
			"version":    "portinfo.version",
			"device":     "portinfo.device",
			"ip":         "ip",
			"port":       "portinfo.port",
			"hostname":   "portinfo.hostname",
			"city":       "geoinfo.city.names.en",
			"city_cn":    "geoinfo.city.names.zh-CN",
			"country":    "geoinfo.country.names.en",
			"country_cn": "geoinfo.country.names.zh-CN",
			"asn":        "geoinfo.asn",
			"banner":     "portinfo.banner",
		},
		"web": map[string]string{
			"_index":     "site",
			"app":        "webapp",
			"headers":    "headers",
			"keywords":   "keywords",
			"title":      "title",
			"ip":         "ip",
			"site":       "site",
			"city":       "geoinfo.city.names.en",
			"city_cn":    "geoinfo.city.names.zh-CN",
			"country":    "geoinfo.country.names.en",
			"country_cn": "geoinfo.country.names.zh-CN",
		},
	}
	statisticsFields = map[string]map[string]string{
		"host": map[string]string{
			"app":     "portinfo.app",
			"device":  "portinfo.device",
			"service": "portinfo.service",
			"os":      "portinfo.os",
			"port":    "portinfo.port",
			"country": "geoinfo.country.names.en",
			"city":    "geoinfo.city.names.en",
		},
		"web": map[string]string{
			"webapp":    "",
			"component": "",
			"framework": "",
			"frontend":  "",
			"server":    "",
			"waf":       "",
			"os":        "",
			"country":   "",
			"city":      "",
		},
	}
)

type findableMap map[string]interface{}

func (m findableMap) Find(expr string) interface{} {
	var (
		set  = strings.Split(expr, ".")
		n    = len(set)
		curr = m
		val  interface{}
		ok   bool
	)
	for i, v := range set {
		if val, ok = curr[v]; !ok {
			return nil
		}
		rc := reflect.ValueOf(val)
		if rc.Kind() != reflect.Map {
			if i < n-1 {
				return nil
			}
			return val
		}
		if vm, ok := val.(map[string]interface{}); ok {
			curr = vm
		} else {
			curr = make(map[string]interface{})
			for _, k := range rc.MapKeys() {
				curr[k.String()] = rc.MapIndex(k).Interface()
			}
		}
	}
	return val
}

func (m findableMap) FindString(expr string) string {
	if v := m.Find(expr); v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// ErrorResult represents result of error
type ErrorResult struct {
	Err     string `json:"error"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

func (r *ErrorResult) Error() string {
	return r.Message
}

// Result represents each type of result
type Result interface {
	setRawData([]byte)
	Raw() []byte
	String() string
}

func toString(r Result) string {
	b, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

type baseResult struct {
	rawData []byte
}

func (r *baseResult) setRawData(raw []byte) {
	r.rawData = raw
}

func (r *baseResult) Raw() []byte {
	return r.rawData
}

// LoginResult represents result of login
type LoginResult struct {
	baseResult
	AccessToken string `json:"access_token"`
}

func (r *LoginResult) String() string {
	return toString(r)
}

// ResourcesInfoResult represents result of resources information
type ResourcesInfoResult struct {
	baseResult
	Plan      string `json:"plan"`
	Resources *struct {
		Interval string `json:"interval"`
		Search   int    `json:"search"`
		Stats    int    `json:"stats"`
	} `json:"resources"`
}

func (r *ResourcesInfoResult) String() string {
	return toString(r)
}

// SearchResult represents result of search
type SearchResult struct {
	baseResult
	Type      string        `json:"-"`
	Available uint64        `json:"available"`
	Total     uint64        `json:"total"`
	Matches   []findableMap `json:"matches"`
	Facets    map[string][]*struct {
		Name  interface{} `json:"name"`
		Count uint64      `json:"count"`
	} `json:"facets"`
	FilterCache []map[string]interface{} `json:"-"`
}

// Sites finds ip and site in web search results
func (r *SearchResult) Sites() []map[string]string {
	m := make([]map[string]string, 0, len(r.Matches))
	for _, v := range r.Matches {
		if ip := v.FindString("ip"); ip != "" {
			m = append(m, map[string]string{
				"ip":   ip,
				"site": v.FindString("site"),
			})
		}
	}
	return m
}

// Hosts finds ip and port in host search results
func (r *SearchResult) Hosts() []map[string]string {
	m := make([]map[string]string, 0, len(r.Matches))
	if r.Type != "host" {

	}
	for _, v := range r.Matches {
		if ip := v.FindString("ip"); ip != "" {
			m = append(m, map[string]string{
				"ip":   ip,
				"port": v.FindString("portinfo.port"),
			})
		}
	}
	return m
}

// Filter extracts data by specified fields from search results
func (r *SearchResult) Filter(keys ...string) []map[string]interface{} {
	var (
		filtered = make([]map[string]interface{}, 0, len(r.Matches))
		n        = len(keys)
	)
	if n == 0 {
		return filtered
	}
	fields, ok := filterFields[r.Type]
	if !ok {
		return filtered
	}
	if n == 1 && keys[0] == "*" {
		keys = make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
	}
	keys = append(keys, "_index")
	for _, v := range r.Matches {
		var (
			item     = make(map[string]interface{})
			count    int
			notMatch bool
		)
		for _, k := range keys {
			var expr string
			if kv := strings.SplitN(k, "=", 2); len(kv) == 2 {
				k = kv[0]
				if expr = strings.TrimSpace(kv[1]); !strings.HasPrefix(expr, "(?i)") {
					expr = "(?i)" + expr
				}
			}
			k = strings.ToLower(strings.TrimSpace(k))
			field, ok := fields[k]
			if !ok {
				continue
			}
			if _, ok = item[k]; ok {
				continue
			}
			find := v.Find(field)
			if k != "_index" && find != nil {
				if expr == "" {
					count++
				} else if reg, err := regexp.Compile(expr); err == nil && reg.MatchString(v.FindString(field)) {
					count++
				} else {
					notMatch = true
					break
				}
			}
			item[k] = find
		}
		if !notMatch && count > 0 {
			filtered = append(filtered, item)
		}
	}
	r.FilterCache = filtered
	return filtered
}

// Statistics counts data by specified fields from search results
func (r *SearchResult) Statistics(keys ...string) map[string]map[string]int {
	var (
		counts     = make(map[string]map[string]int)
		fields, ok = statisticsFields[r.Type]
	)
	if !ok {
		return counts
	}
	for _, k := range keys {
		k = strings.ToLower(strings.TrimSpace(k))
		field, ok := fields[k]
		if !ok {
			continue
		}
		if _, ok := counts[k]; !ok {
			counts[k] = make(map[string]int)
		}
		for _, v := range r.Matches {
			name := v.FindString(field)
			if name == "" {
				name = "[unknown]"
			}
			counts[k][name]++
		}
	}
	return counts
}

// Extend merges more than one search results
func (r *SearchResult) Extend(res *SearchResult) {
	if res == nil {
		return
	}
	if res.Type != "" {
		if r.Type == "" {
			r.Type = res.Type
		} else if r.Type != res.Type {
			return
		}
	}
	if n := len(r.rawData); n == 0 {
		r.rawData = res.rawData
	} else if r.rawData[0] == '[' && r.rawData[n-1] == ']' {
		r.rawData = append(r.rawData[:n-1], make([]byte, len(res.rawData)+3)...)
		copy(r.rawData[n-1:], []byte(", "))
		copy(r.rawData[n+1:], res.rawData)
		r.rawData[len(r.rawData)-1] = ']'
	} else {
		raw := r.rawData
		r.rawData = make([]byte, len(res.rawData)+n+4)
		r.rawData[0] = '['
		copy(r.rawData[1:], raw)
		copy(r.rawData[n+1:], []byte(", "))
		copy(r.rawData[n+3:], res.rawData)
		r.rawData[len(r.rawData)-1] = ']'
	}
	if res.Total > 0 {
		r.Available = res.Available
		r.Total = res.Total
		r.Matches = append(r.Matches, res.Matches...)
		r.Facets = res.Facets
	}
}

func (r *SearchResult) String() string {
	return toString(r)
}

// HistoryResult represents result of history
type HistoryResult struct {
	baseResult
	Count uint64        `json:"count"`
	Data  []findableMap `json:"data"`
}

func (r *HistoryResult) String() string {
	return toString(r)
}
