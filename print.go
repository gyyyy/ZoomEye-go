package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gyyyy/ZoomEye-go/zoomeye"
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
	ctrlChars = map[rune]string{
		'\t': "\\t",
		'\n': "\\n",
		'\v': "\\v",
		'\f': "\\f",
		'\r': "\\r",
		'\a': "\\a",
		'\b': "\\b",
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
	print(fmt.Sprintf(format, a...), colorRed)
}

func successf(format string, a ...interface{}) {
	print(fmt.Sprintf(format, a...), colorGreen)
}

func warnf(format string, a ...interface{}) {
	print(fmt.Sprintf(format, a...), colorYellow)
}

func infof(title, format string, a ...interface{}) {
	if title != "" {
		format = "\n" + colorf("["+title+"]", colorLightCyan) + "\n\n" +
			colorf("  "+strings.ReplaceAll(format, "\n", "\n  "), colorLightWhite) + "\n"
		print(fmt.Sprintf(format, a...), "")
	} else {
		print(fmt.Sprintf(format, a...), colorWhite)
	}
}

func toStr(o interface{}) string {
	if o == nil {
		return ""
	}
	switch o := o.(type) {
	case string:
		return o
	case []string:
		return strings.Join(o, ",")
	case []interface{}:
		s := make([]string, len(o))
		for i, v := range o {
			s[i] = fmt.Sprintf("%v", v)
		}
		return strings.Join(s, ",")
	default:
		return fmt.Sprintf("%v", o)
	}
}

func omitStr(o interface{}, maxWidth int) string {
	var builder strings.Builder
	for _, r := range toStr(o) {
		if r > 31 && r < 127 {
			builder.WriteRune(r)
		} else if v, ok := ctrlChars[r]; ok {
			builder.WriteString(v)
		} else {
			builder.WriteString(fmt.Sprintf("\\x%02x", r))
		}
	}
	s := builder.String()
	if n := len(s); n > maxWidth {
		if maxWidth > 3 {
			s = s[:maxWidth-3] + "..."
		} else {
			s = strings.Repeat(".", maxWidth)
		}
	}
	return strings.ReplaceAll(s, "%", "%%")
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
			if isGroup {
				k = omitStr(k, widths[0])
			}
			builder.WriteString(line + "\n")
			for i, v := range group {
				if len(v) < len(head)-1 {
					continue
				}
				v = v[:len(head)-1]
				for i := range v {
					j := i
					if isGroup {
						j++
					}
					v[i] = omitStr(v[i], widths[j])
				}
				if total++; k == "" && !isGroup {
					builder.WriteString(fmt.Sprintf(bfmt, v...) + "\n")
					continue
				}
				if i > 0 {
					k = ""
				}
				params := append([]interface{}{k}, v...)
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
				colorf("|", colorLightBlack)+
					colorf(" %%-%ds ", colorLightPurple)+
					colorf("|", colorLightBlack)+"\n",
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
			name := omitStr(item["name"], widths[0])
			for i, v := range item["items"].([]map[string]interface{}) {
				if i > 0 {
					name = ""
				}
				var (
					key = omitStr(v["key"], widths[1])
					val = omitStr(v["value"], widths[2])
				)
				builder.WriteString(fmt.Sprintf(bfmt, name, key, val) + "\n")
			}
		}
	} else {
		builder.WriteString(line + "\n")
	}
	if count {
		builder.WriteString(line + "\n")
		builder.WriteString(fmt.Sprintf(
			fmt.Sprintf(
				colorf("|", colorLightBlack)+
					colorf(" %%-%ds ", colorLightPurple)+
					colorf("|", colorLightBlack)+"\n",
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
						builder.WriteString(c + "   " + colorf("Type: "+k, colorLightGreen) + "\n")
					} else if n := i - 3; n >= 0 && n < len(v) {
						name := omitStr(v[n][0], 35)
						builder.WriteString(c + "   " +
							colorf(fmt.Sprintf("%5.2f%%%% - %s", v[n][2].(float64)*100, name), pieColors[n]) + "\n")
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
			builder.WriteString(colorf("Type: "+k, colorLightGreen) + "\n\n")
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
			if maxNameLen > 35 {
				maxNameLen = 35
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
				builder.WriteString(colorf(fmt.Sprintf(format, omitStr(o[0], 35), o[1], bar), colorLightWhite))
				if i < len(v)-1 {
					builder.WriteString("\n")
				}
			}
		}
	}
	infof(title, builder.String())
}

func withUnknown(o interface{}) string {
	if s := toStr(o); s != "" && !strings.EqualFold(strings.TrimSpace(s), "unknown") {
		return s
	}
	return "[unknown]"
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
				s += v["name"].(string)
				if ver := v["version"]; ver != nil && ver != "" {
					s += "(" + ver.(string) + ")"
				}
			}
		}
		return s
	}
	return ""
}

func showFacet(result *zoomeye.SearchResult, facets []string, figure string) {
	var (
		head = [][2]interface{}{
			{"Type", 10},
			{"Name", 35},
			{"Count", 20},
		}
		body = make(map[string][][]interface{})
	)
	for _, f := range facets {
		f = strings.ToLower(strings.TrimSpace(f))
		s := f
		if result.Type == "host" && f == "app" {
			s = "product"
		}
		if facet, ok := result.Facets[s]; ok {
			group := make([][]interface{}, 0, len(facet))
			for _, v := range facet {
				group = append(group, []interface{}{
					withUnknown(v.Name),
					v.Count,
					float64(v.Count) / float64(result.Total),
				})
			}
			body[f] = group
		}
	}
	switch figure {
	case "":
		tablef("ZoomEye Facets", head, body, false)
	case "pie":
		pief("ZoomEye Facets - PIE", body)
	case "hist":
		histf("ZoomEye Facets - HIST", body)
	}
}

func showStat(result *zoomeye.SearchResult, keys []string, figure string) {
	var (
		head = [][2]interface{}{
			{"Type", 10},
			{"Name", 35},
			{"Count", 20},
		}
		body = make(map[string][][]interface{})
	)
	for s, stat := range result.Statistics(keys...) {
		group := make([][]interface{}, 0, len(stat))
		for k, v := range stat {
			group = append(group, []interface{}{
				k,
				v,
				float64(v) / float64(len(result.Matches)),
			})
		}
		sort.Slice(group, func(i, j int) bool {
			return group[j][1].(uint64) < group[i][1].(uint64)
		})
		body[s] = group
	}
	switch figure {
	case "":
		tablef("Result Statistics", head, body, false)
	case "pie":
		pief("Result Statistics - PIE", body)
	case "hist":
		histf("Result Statistics - HIST", body)
	}
}

func showFilter(result *zoomeye.SearchResult, keys []string) []map[string]interface{} {
	var (
		filtered = result.Filter(keys...)
		body     = make([]map[string]interface{}, len(filtered))
	)
	for i, filt := range filtered {
		var (
			index = filt["_index"].(string)
			items = make([]map[string]interface{}, 0, len(filt)-1)
		)
		for k, v := range filt {
			if k != "_index" {
				items = append(items, map[string]interface{}{
					"key":   strings.ToTitle(k),
					"value": v,
				})
			}
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i]["key"].(string) < items[j]["key"].(string)
		})
		body[i] = map[string]interface{}{
			"name":  index,
			"items": items,
		}
	}
	htablef("Result Filtered", body, [3]int{30, 15, 75}, true)
	return filtered
}

func showData(result *zoomeye.SearchResult) {
	switch result.Type {
	case "host":
		var (
			head = [][2]interface{}{
				{"-", 0},
				{"Host", 21},
				{"Application", 20},
				{"Service", 20},
				{"Banner", 40},
				{"Country", 20},
			}
			body = make([][]interface{}, len(result.Matches))
		)
		for i, v := range result.Matches {
			body[i] = []interface{}{
				v.FindString("ip") + ":" + v.FindString("portinfo.port"),
				v.FindString("portinfo.app"),
				v.FindString("portinfo.service"),
				v.FindString("portinfo.banner"),
				v.FindString("geoinfo.country.names.en"),
			}
		}
		tablef("Host Search Result", head, map[string][][]interface{}{"": body}, true)
	case "web":
		body := make([]map[string]interface{}, len(result.Matches))
		for i, v := range result.Matches {
			body[i] = map[string]interface{}{
				"name": v.FindString("site"),
				"items": []map[string]interface{}{
					{
						"key":   "IP",
						"value": v.Find("ip"),
					},
					{
						"key":   "Domains",
						"value": v.Find("domains"),
					},
					{
						"key":   "Country",
						"value": v.FindString("geoinfo.country.names.en"),
					},
					{
						"key":   "Title",
						"value": v.FindString("title"),
					},
					{
						"key":   "Application",
						"value": withVersion(v.Find("webapp")),
					},
					{
						"key":   "Framework",
						"value": withVersion(v.Find("framework")),
					},
					{
						"key":   "Server",
						"value": withVersion(v.Find("server")),
					},
					{
						"key":   "System",
						"value": withVersion(v.Find("system")),
					},
					{
						"key":   "Database",
						"value": withVersion(v.Find("db")),
					},
					{
						"key":   "WAF",
						"value": withVersion(v.Find("waf")),
					},
				},
			}
		}
		htablef("Web Search Result", body, [3]int{30, 15, 75}, true)
	}
}

func showHistory(result *zoomeye.HistoryResult, keys []string, num int) {
	var (
		filtered = result.Filter(keys...)
		n        = len(filtered)
	)
	if n == 0 {
		infof("[History Info]", "no any historical data")
		return
	}
	if num > 0 && num < n {
		filtered = filtered[:n]
	}
	var (
		first = filtered[0]
		info  = fmt.Sprintf("%s\n\n"+
			"Hostname:          %s\n"+
			"Country:           %s\n"+
			"City:              %s\n"+
			"Organization:      %s\n"+
			"Last Updated:      %s\n\n",
			first["ip"], withUnknown(first["host"]), withUnknown(first["country"]),
			withUnknown(first["city"]), withUnknown(first["org"]), first["last_update"])
	)
	var (
		head = [][2]interface{}{
			{"-", 0},
			{"Time", 19},
			{"Port", 5},
			{"Service", 25},
			{"App", 25},
			{"Raw", 45},
		}
		body = make([][]interface{}, len(filtered))
	)
	for i := 2; i < len(head); {
		if _, ok := first[strings.ToLower(head[i][0].(string))]; !ok {
			head = append(head[:i], head[i+1:]...)
		} else {
			i++
		}
	}
	if len(head) == 6 {
		head[3][0] = "Port/Service"
		head = append(head[:2], head[3:]...)
	}
	ports := make(map[string]struct{})
	for i, f := range filtered {
		if p := toStr(f["open_port"]); p != "" {
			ports[p] = struct{}{}
		}
		row := make([]interface{}, 0, 5)
		for i := 1; i < len(head); i++ {
			var (
				sp = strings.SplitN(head[i][0].(string), "/", 2)
				v  = make([]string, 0, 2)
			)
			for _, k := range sp {
				v = append(v, toStr(f[strings.ToLower(k)]))
			}
			row = append(row, strings.Join(v, "/"))
		}
		if row[0] == "" {
			row[0] = f["last_update"]
		}
		body[i] = row
	}
	info += fmt.Sprintf("Open Ports:        %d\nHistorical Probes: %d", len(ports), len(filtered))
	infof("History Info", info)
	tablef("History Result", head, map[string][][]interface{}{"": body}, true)
}
