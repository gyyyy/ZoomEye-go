package main

import (
	"ZoomEye-go/zoomeye"
	"fmt"
	"strings"
)

var (
	spaces = map[rune]string{
		'\t': "\\t",
		'\n': "\\n",
		'\v': "\\v",
		'\f': "\\f",
		'\r': "\\r",
	}
	colors = map[string]string{
		"red":    "\033[91m",
		"green":  "\033[92m",
		"yellow": "\033[93m",
		"white":  "\033[97m",
	}
)

func print(s, color string) {
	c, ok := colors[color]
	if !ok {
		c = colors["white"]
	}
	fmt.Printf("%s%s\033[0m\n", c, s)
}

func errorf(format string, a ...interface{}) {
	print(fmt.Sprintf(format, a...), "red")
}

func successf(format string, a ...interface{}) {
	print(fmt.Sprintf(format, a...), "green")
}

func warnf(format string, a ...interface{}) {
	print(fmt.Sprintf(format, a...), "yellow")
}

func infof(title, format string, a ...interface{}) {
	if format = strings.ReplaceAll(format, "\n", "\n  "); title != "" {
		format = "\n[" + title + "]\n\n  " + format + "\n"
	}
	print(fmt.Sprintf(format, a...), "")
}

func tablef(title string, head [][2]interface{}, body map[string][][]interface{}, count bool) {
	var (
		builder strings.Builder
		n       = len(head)
		names   = make([]interface{}, 0, n)
		widths  = make([]int, 0, n)
		format  = "|"
		isGroup bool
	)
	for i, v := range head {
		var (
			name  = v[0].(string)
			width = v[1].(int)
		)
		if i == 0 && name != "-" {
			isGroup = true
		}
		if i > 0 || name != "-" {
			format += fmt.Sprintf(" %%-%dv |", width)
			names = append(names, name)
			widths = append(widths, width)
		}
	}
	line := "+"
	for _, v := range widths {
		line += strings.Repeat("-", v+2) + "+"
	}
	var total int
	builder.WriteString(line + "\n")
	builder.WriteString(fmt.Sprintf(format, names...) + "\n")
	if len(body) > 0 {
		for k, group := range body {
			builder.WriteString(line + "\n")
			for i, v := range group {
				if total++; k == "" && !isGroup {
					builder.WriteString(fmt.Sprintf(format, v...) + "\n")
					continue
				}
				if i > 0 {
					k = ""
				}
				params := append([]interface{}{k}, v...)
				builder.WriteString(fmt.Sprintf(format, params...) + "\n")
			}
		}
	} else {
		builder.WriteString(line + "\n")
	}
	if count {
		builder.WriteString(line + "\n")
		builder.WriteString(fmt.Sprintf(fmt.Sprintf("| %%-%ds |\n", len(line)-4), fmt.Sprintf("Total: %d", total)))
	}
	builder.WriteString(line)
	infof(title, builder.String())
}

func htablef(title string, body []map[string]interface{}, widths [3]int, count bool) {
	var (
		builder strings.Builder
		format  = fmt.Sprintf("| %%-%dv | %%-%dv | %%-%dv |", widths[0], widths[1], widths[2])
		line    = "+"
	)
	for _, v := range widths {
		line += strings.Repeat("-", v+2) + "+"
	}
	var total int
	builder.WriteString(line + "\n")
	builder.WriteString(fmt.Sprintf(format, "Name", "Key", "Value") + "\n")
	if len(body) > 0 {
		for _, item := range body {
			total++
			builder.WriteString(line + "\n")
			name := item["name"].(string)
			for i, v := range item["items"].([]map[string]interface{}) {
				if i > 0 {
					name = ""
				}
				builder.WriteString(fmt.Sprintf(format, name, v["key"], v["value"]) + "\n")
			}
		}
	} else {
		builder.WriteString(line + "\n")
	}
	if count {
		builder.WriteString(line + "\n")
		builder.WriteString(fmt.Sprintf(fmt.Sprintf("| %%-%ds |\n", len(line)-4), fmt.Sprintf("Total: %d", total)))
	}
	builder.WriteString(line)
	infof(title, builder.String())
}

func convertStr(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if r > 31 && r < 127 {
			builder.WriteRune(r)
		} else if v, ok := spaces[r]; ok {
			builder.WriteString(v)
		} else {
			builder.WriteString(fmt.Sprintf("\\x%02x", r))
		}
	}
	return builder.String()
}

func omitStr(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}
	return s[:maxWidth-3] + "..."
}

func printFacet(result *zoomeye.SearchResult, facets []string) {
	var (
		head = [][2]interface{}{
			[2]interface{}{"Type", 10},
			[2]interface{}{"Facet", 35},
			[2]interface{}{"Count", 20},
		}
		body = make(map[string][][]interface{})
	)
	for _, s := range facets {
		if s = strings.ToLower(strings.TrimSpace(s)); result.Type == "host" && s == "app" {
			s = "product"
		}
		if facet, ok := result.Facets[s]; ok {
			group := make([][]interface{}, 0)
			for _, v := range facet {
				var name string
				switch n := v.Name.(type) {
				case string:
					name = n
				case nil:
				default:
					name = fmt.Sprintf("%v", n)
				}
				if name == "" {
					name = "[unknown]"
				}
				group = append(group, []interface{}{omitStr(name, 35), v.Count})
			}
			body[s] = group
		}
	}
	tablef("Facets Info", head, body, false)
}

func printStat(result *zoomeye.SearchResult, keys []string) {
	var (
		head = [][2]interface{}{
			[2]interface{}{"Type", 10},
			[2]interface{}{"Name", 35},
			[2]interface{}{"Count", 20},
		}
		body = make(map[string][][]interface{})
	)
	for s, stat := range result.Statistics(keys...) {
		group := make([][]interface{}, 0)
		for k, v := range stat {
			group = append(group, []interface{}{omitStr(k, 35), v})
		}
		body[s] = group
	}
	tablef("Statistics Info", head, body, false)
}

func dataToString(o interface{}) string {
	if o == nil {
		return ""
	}
	switch o := o.(type) {
	case string:
		return o
	case []string:
		return strings.Join(o, ",")
	case []interface{}:
		var s string
		for i, v := range o {
			if i > 0 {
				s += ","
			}
			s += fmt.Sprintf("%v", v)
		}
		return s
	default:
		return fmt.Sprintf("%v", o)
	}
}

func printFilter(result *zoomeye.SearchResult, keys []string) {
	var (
		filters = result.Filter(keys...)
		body    = make([]map[string]interface{}, len(filters))
	)
	for i, filt := range filters {
		index := filt["index"].(string)
		delete(filt, "index")
		items := make([]map[string]interface{}, 0, len(filt))
		for k, v := range filt {
			items = append(items, map[string]interface{}{
				"key":   strings.ToTitle(k),
				"value": omitStr(convertStr(dataToString(v)), 75),
			})
		}
		body[i] = map[string]interface{}{
			"name":  omitStr(index, 30),
			"items": items,
		}
	}
	htablef("Filtered Data", body, [3]int{30, 15, 75}, true)
}

func withVersion(o interface{}) string {
	if o == nil {
		return ""
	}
	if o, ok := o.([]interface{}); ok {
		var s string
		for i, v := range o {
			if v, ok := v.(map[string]interface{}); ok {
				if i > 0 {
					s += ","
				}
				var (
					name = v["name"]
					ver  = v["version"]
				)
				s += name.(string)
				if ver != nil && ver != "" {
					s += "(" + ver.(string) + ")"
				}
			}
		}
		return s
	}
	return ""
}

func printData(result *zoomeye.SearchResult) {
	switch result.Type {
	case "host":
		var (
			head = [][2]interface{}{
				[2]interface{}{"-", 0},
				[2]interface{}{"Host", 21},
				[2]interface{}{"Application", 20},
				[2]interface{}{"Service", 20},
				[2]interface{}{"Banner", 40},
				[2]interface{}{"Country", 20},
			}
			body = make([][]interface{}, len(result.Matches))
		)
		for i, v := range result.Matches {
			body[i] = []interface{}{
				v.FindString("ip") + ":" + v.FindString("portinfo.port"),
				omitStr(v.FindString("portinfo.app"), 20),
				omitStr(v.FindString("portinfo.service"), 20),
				omitStr(convertStr(v.FindString("portinfo.banner")), 40),
				omitStr(v.FindString("geoinfo.country.names.en"), 20),
			}
		}
		tablef("Host Search Result", head, map[string][][]interface{}{"": body}, true)
	case "web":
		body := make([]map[string]interface{}, len(result.Matches))
		for i, v := range result.Matches {
			body[i] = map[string]interface{}{
				"name": omitStr(v.FindString("site"), 30),
				"items": []map[string]interface{}{
					map[string]interface{}{
						"key":   "IP",
						"value": omitStr(dataToString(v.Find("ip")), 75),
					},
					map[string]interface{}{
						"key":   "Domains",
						"value": omitStr(dataToString(v.Find("domains")), 75),
					},
					map[string]interface{}{
						"key":   "Application",
						"value": omitStr(withVersion(v.Find("webapp")), 75),
					},
					map[string]interface{}{
						"key":   "Title",
						"value": omitStr(convertStr(v.FindString("title")), 75),
					},
					map[string]interface{}{
						"key":   "Framework",
						"value": omitStr(withVersion(v.Find("framework")), 75),
					},
					map[string]interface{}{
						"key":   "Server",
						"value": omitStr(withVersion(v.Find("server")), 75),
					},
					map[string]interface{}{
						"key":   "System",
						"value": omitStr(withVersion(v.Find("system")), 75),
					},
					map[string]interface{}{
						"key":   "Database",
						"value": omitStr(withVersion(v.Find("db")), 75),
					},
					map[string]interface{}{
						"key":   "WAF",
						"value": omitStr(withVersion(v.Find("waf")), 75),
					},
					map[string]interface{}{
						"key":   "Country",
						"value": v.FindString("geoinfo.country.names.en"),
					},
				},
			}
		}
		htablef("Web Search Result", body, [3]int{30, 15, 75}, true)
	}
}
