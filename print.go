package main

import (
	"ZoomEye-go/zoomeye"
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	colorReset       = "\033[0m"
	colorBlack       = "\033[0;30m"
	colorRed         = "\033[0;31m"
	colorGreen       = "\033[0;32m"
	colorYellow      = "\033[0;33m"
	colorBlue        = "\033[0;34m"
	colorPurple      = "\033[0;35m"
	colorCyan        = "\033[0;36m"
	colorWhite       = "\033[0;37m"
	colorLightBlack  = "\033[1;30m"
	colorLightRed    = "\033[1;31m"
	colorLightGreen  = "\033[1;32m"
	colorLightYellow = "\033[1;33m"
	colorLightBlue   = "\033[1;34m"
	colorLightPurple = "\033[1;35m"
	colorLightCyan   = "\033[1;36m"
	colorLightWhite  = "\033[1;37m"
	colorDarkBlack   = "\033[2;30m"
	colorDarkRed     = "\033[2;31m"
	colorDarkGreen   = "\033[2;32m"
	colorDarkYellow  = "\033[2;33m"
	colorDarkBlue    = "\033[2;34m"
	colorDarkPurple  = "\033[2;35m"
	colorDarkCyan    = "\033[2;36m"
	colorDarkWhite   = "\033[2;37m"
)

var (
	spaces = map[rune]string{
		'\t': "\\t",
		'\n': "\\n",
		'\v': "\\v",
		'\f': "\\f",
		'\r': "\\r",
	}
	pieColors = []string{
		"\033[1;34m", "\033[1;35m", "\033[1;36m", "\033[1;31m", "\033[1;33m",
		"\033[0;94m", "\033[0;95m", "\033[0;96m", "\033[0;91m", "\033[0;93m",
	}
	histChars = []string{
		" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█",
	}
)

func colorf(s, color string) string {
	if color == "" {
		return s
	}
	return color + s + colorReset
}

func print(s, color string) {
	fmt.Println(colorf(s, color))
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
	if title != "" {
		format = "\n" + colorf("["+title+"]", colorLightCyan) + "\n\n" +
			colorf("  "+strings.ReplaceAll(format, "\n", "\n  "), colorLightWhite) + "\n"
	}
	print(fmt.Sprintf(format, a...), "")
}

func tablef(title string, head [][2]interface{}, body map[string][][]interface{}, count bool) {
	var (
		builder strings.Builder
		n       = len(head)
		names   = make([]interface{}, 0, n)
		widths  = make([]int, 0, n)
		hfmt    = colorf("|", colorLightBlack)
		bfmt    = colorf("|", colorLightBlack)
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
			hfmt += fmt.Sprintf(colorf(" %%-%dv ", colorLightGreen)+colorf("|", colorLightBlack), width)
			bfmt += fmt.Sprintf(colorf(" %%-%dv ", colorLightWhite)+colorf("|", colorLightBlack), width)
			names = append(names, name)
			widths = append(widths, width)
		}
	}
	line := "+"
	for _, v := range widths {
		line += strings.Repeat("-", v+2) + "+"
	}
	line = colorf(line, colorLightBlack)
	var total int
	builder.WriteString(line + "\n")
	builder.WriteString(fmt.Sprintf(hfmt, names...) + "\n")
	if len(body) > 0 {
		for k, group := range body {
			builder.WriteString(line + "\n")
			for i, v := range group {
				if len(v) < len(head)-1 {
					continue
				}
				if total++; k == "" && !isGroup {
					builder.WriteString(fmt.Sprintf(bfmt, v[:len(head)-1]...) + "\n")
					continue
				}
				if i > 0 {
					k = ""
				}
				params := append([]interface{}{k}, v[:len(head)-1]...)
				builder.WriteString(fmt.Sprintf(bfmt, params...) + "\n")
			}
		}
	} else {
		builder.WriteString(line + "\n")
	}
	if count {
		builder.WriteString(line + "\n")
		builder.WriteString(fmt.Sprintf(
			fmt.Sprintf(
				colorf("|", colorLightBlack)+colorf(" %%-%ds ", colorLightPurple)+colorf("|", colorLightBlack)+"\n",
				len(line)-15,
			),
			fmt.Sprintf("Total: %d", total),
		))
	}
	builder.WriteString(line)
	infof(title, builder.String())
}

func htablef(title string, body []map[string]interface{}, widths [3]int, count bool) {
	var (
		builder strings.Builder
		hfmt    = fmt.Sprintf(colorf("|", colorLightBlack)+
			colorf(" %%-%dv ", colorLightGreen)+colorf("|", colorLightBlack)+
			colorf(" %%-%dv ", colorLightGreen)+colorf("|", colorLightBlack)+
			colorf(" %%-%dv ", colorLightGreen)+colorf("|", colorLightBlack),
			widths[0], widths[1], widths[2])
		bfmt = fmt.Sprintf(colorf("|", colorLightBlack)+
			colorf(" %%-%dv ", colorLightWhite)+colorf("|", colorLightBlack)+
			colorf(" %%-%dv ", colorLightWhite)+colorf("|", colorLightBlack)+
			colorf(" %%-%dv ", colorLightWhite)+colorf("|", colorLightBlack),
			widths[0], widths[1], widths[2])
		line = "+"
	)
	for _, v := range widths {
		line += strings.Repeat("-", v+2) + "+"
	}
	line = colorf(line, colorLightBlack)
	var total int
	builder.WriteString(line + "\n")
	builder.WriteString(fmt.Sprintf(hfmt, "Name", "Key", "Value") + "\n")
	if len(body) > 0 {
		for _, item := range body {
			total++
			builder.WriteString(line + "\n")
			name := item["name"].(string)
			for i, v := range item["items"].([]map[string]interface{}) {
				if i > 0 {
					name = ""
				}
				builder.WriteString(fmt.Sprintf(bfmt, name, v["key"], v["value"]) + "\n")
			}
		}
	} else {
		builder.WriteString(line + "\n")
	}
	if count {
		builder.WriteString(line + "\n")
		builder.WriteString(fmt.Sprintf(
			fmt.Sprintf(
				colorf("|", colorLightBlack)+colorf(" %%-%ds ", colorLightPurple)+colorf("|", colorLightBlack)+"\n",
				len(line)-15,
			),
			fmt.Sprintf("Total: %d", total),
		))
	}
	builder.WriteString(line)
	infof(title, builder.String())
}

func atanChar(data [][]interface{}, at float64, colors []string) string {
	if len(data) == 0 {
		return "  "
	}
	n := at - data[0][2].(float64)
	if n <= 0 {
		return colors[0] + "* " + colorReset
	}
	return atanChar(data[1:], n, colors[1:])
}

func pief(title string, body map[string][][]interface{}) {
	var builder strings.Builder
	if len(body) > 0 {
		first := true
		for k, v := range body {
			if !first {
				builder.WriteString("\n\n\n")
			} else {
				first = false
			}
			if len(v) > 10 {
				v = v[:10]
			}
			for i, y := 0, -7; y < 7; y++ {
				var c string
				for x := -7; x < 7; x++ {
					if x*x+y*y < 49 {
						c += atanChar(v, math.Atan2(float64(y), float64(x))/math.Pi/2+0.5, pieColors)
					} else {
						c += "  "
					}
				}
				if i > 0 {
					if i == 1 {
						builder.WriteString(c + "   " + colorf(strings.ToUpper(k), colorLightGreen) + "\n")
					} else if n := i - 3; n >= 0 && n < len(v) {
						builder.WriteString(c + "   " +
							colorf(fmt.Sprintf("%5.2f%%%% - %s", v[n][2].(float64)*100, v[n][0]), pieColors[n]) + "\n")
					} else if builder.WriteString(c); i < 13 {
						builder.WriteString("\n")
					}
				}
				i++
			}
		}
	}
	infof(title, builder.String())
}

func histf(title string, body map[string][][]interface{}) {
	var builder strings.Builder
	if len(body) > 0 {
		first := true
		for k, v := range body {
			if !first {
				builder.WriteString("\n\n\n")
			} else {
				first = false
			}
			builder.WriteString(colorf(strings.ToUpper(k), colorLightGreen) + "\n\n")
			var (
				maxNameLen  int
				maxCountLen int
				maxCount    uint64
			)
			for _, o := range v {
				if n := len(o[0].(string)); n > maxNameLen {
					maxNameLen = n
				}
				if n := len(fmt.Sprintf("%d", o[1])); n > maxCountLen {
					maxCountLen = n
				}
				if n := o[1].(uint64); n > maxCount {
					maxCount = n
				}
			}
			format := fmt.Sprintf("%%%ds  [%%%dd]  %%s", maxNameLen, maxCountLen)
			for i, o := range v {
				var (
					n   = int(math.Round(float64(o[1].(uint64)) / float64(maxCount) * 36 * 8))
					bar = strings.Repeat(histChars[7], n/8)
				)
				if n%8 > 0 {
					bar += histChars[n%8]
				}
				builder.WriteString(colorf(fmt.Sprintf(format, o[0], o[1], bar), colorLightWhite))
				if i < len(v)-1 {
					builder.WriteString("\n")
				}
			}
		}
	}
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
	if len(s) > maxWidth {
		s = s[:maxWidth-3] + "..."
	}
	return strings.ReplaceAll(s, "%", "%%")
}

func printFacet(result *zoomeye.SearchResult, facets []string, figure string) {
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
			group := make([][]interface{}, 0, len(facet))
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
				group = append(group, []interface{}{omitStr(name, 35), v.Count,
					float64(v.Count) / float64(result.Total)})
			}
			body[s] = group
		}
	}
	switch figure {
	case "":
		tablef("Facets Info", head, body, false)
	case "pie":
		pief("Facets Pie", body)
	case "hist":
		histf("Facets Histogram", body)
	}
}

func printStat(result *zoomeye.SearchResult, keys []string, figure string) {
	var (
		head = [][2]interface{}{
			[2]interface{}{"Type", 10},
			[2]interface{}{"Name", 35},
			[2]interface{}{"Count", 20},
		}
		body = make(map[string][][]interface{})
	)
	for s, stat := range result.Statistics(keys...) {
		group := make([][]interface{}, 0, len(stat))
		for k, v := range stat {
			group = append(group, []interface{}{omitStr(k, 35), v, float64(v) / float64(len(result.Matches))})
		}
		sort.Slice(group, func(i, j int) bool {
			return group[j][1].(uint64) < group[i][1].(uint64)
		})
		body[s] = group
	}
	switch figure {
	case "":
		tablef("Statistics Info", head, body, false)
	case "pie":
		pief("Statistics Pie", body)
	case "hist":
		histf("Statistics Histogram", body)
	}
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
		var (
			index = filt["_index"].(string)
			items = make([]map[string]interface{}, 0, len(filt)-1)
		)
		for k, v := range filt {
			if k != "_index" {
				items = append(items, map[string]interface{}{
					"key":   strings.ToTitle(k),
					"value": omitStr(convertStr(dataToString(v)), 75),
				})
			}
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
				omitStr(v.FindString("ip")+":"+v.FindString("portinfo.port"), 21),
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
