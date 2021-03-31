package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/gyyyy/ZoomEye-go/zoomeye"
)

func parseFlags(cmd string, flgs interface{}, examples ...string) []string {
	if flgs != nil {
		var (
			v    = reflect.ValueOf(flgs)
			elem = v.Elem()
			ptr  = v.Pointer()
		)
		for i := 0; i < elem.NumField(); i++ {
			var (
				f      = elem.Type().Field(i)
				fname  = f.Tag.Get("name")
				fvalue = f.Tag.Get("value")
				fusage = f.Tag.Get("usage")
				fptr   = unsafe.Pointer(ptr + f.Offset)
			)
			if fname == "" {
				fname = strings.ToLower(f.Name)
			}
			switch f.Type.Kind() {
			case reflect.String:
				flag.StringVar((*string)(fptr), fname, fvalue, fusage)
			case reflect.Int:
				val, _ := strconv.Atoi(fvalue)
				flag.IntVar((*int)(fptr), fname, val, fusage)
			case reflect.Bool:
				val, _ := strconv.ParseBool(fvalue)
				flag.BoolVar((*bool)(fptr), fname, val, fusage)
			}
		}
	}
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s (%s):\n", filepath.Base(os.Args[0]), cmd)
		flag.PrintDefaults()
		if len(examples) > 0 {
			example := "Example:\n"
			for _, v := range examples {
				example += fmt.Sprintf("  ./ZoomEye-go %s %s\n", cmd, v)
			}
			fmt.Fprintf(flag.CommandLine.Output(), "\n%s\n", example)
		}
	}
	var args []string
	for i := 1; i < len(os.Args); {
		if p := os.Args[i]; !strings.HasPrefix(p, "-") {
			args = append(args, p)
			os.Args = append(os.Args[0:i], os.Args[i+1:]...)
		} else {
			break
		}
	}
	flag.Parse()
	return args
}

func checkError(err error) {
	switch err.(type) {
	case *NoAuthKeyErr:
		warnf("not found any Auth Keys, please run <zoomeye init> first")
	case *zoomeye.ErrorResult:
		errorf("failed to authenticate: %v", err)
	case nil:
	default:
		errorf("something is wrong: %v", err)
	}
}

type resultAnalyzer struct {
	count  bool
	facet  string
	stat   string
	figure string
	filter string
	save   bool
}

func newResultAnalyzer() *resultAnalyzer {
	analyzer := &resultAnalyzer{}
	flag.BoolVar(&analyzer.count, "count", false, "The total number of results in ZoomEye database")
	flag.StringVar(&analyzer.facet, "facet", "", "Perform statistics on ZoomEye database")
	flag.StringVar(&analyzer.stat, "stat", "", "Perform statistics on search results")
	flag.StringVar(&analyzer.figure, "figure", "", "Output Pie or bar chart only be used under -facet and -stat")
	flag.StringVar(&analyzer.filter, "filter", "", "Output more clearer search results by set filter field")
	flag.BoolVar(&analyzer.save, "save", false, "Save data in JSON format")
	return analyzer
}

func (a *resultAnalyzer) do(result *zoomeye.SearchResult, saveCallback func([]map[string]interface{})) {
	if a.count {
		infof("ZoomEye Total", "Count: %d", result.Total)
	}
	if a.figure != "" {
		if a.figure = strings.ToLower(a.figure); a.figure != "pie" {
			a.figure = "hist"
		}
	}
	if a.facet != "" {
		showFacet(result, strings.Split(a.facet, ","), a.figure)
	}
	if a.stat != "" {
		showStat(result, strings.Split(a.stat, ","), a.figure)
	}
	var filtered []map[string]interface{}
	if a.filter != "" {
		filtered = showFilter(result, strings.Split(a.filter, ","))
	}
	if !a.count && a.facet == "" && a.stat == "" && a.filter == "" {
		showData(result)
	}
	if a.save && saveCallback != nil {
		saveCallback(filtered)
	}
}

func cmdInit(agent *ZoomEyeAgent) {
	var flgs struct {
		apiKey   string `usage:"ZoomEye API-Key"`
		username string `usage:"ZoomEye account username"`
		password string `usage:"ZoomEye account password"`
	}
	parseFlags("init", &flgs, `-apikey "XXXXXXXX-XXXX-XXXXX-XXXX-XXXXXXXXXXX"`,
		`-username "username@zoomeye.org" -password "password"`)
	var (
		result *zoomeye.ResourcesInfoResult
		err    error
	)
	if flgs.apiKey != "" {
		if result, err = agent.InitByKey(flgs.apiKey); err != nil {
			errorf("failed to initialize: %v", err)
			return
		}
	} else if flgs.username != "" && flgs.password != "" {
		if result, err = agent.InitByUser(flgs.username, flgs.password); err != nil {
			errorf("failed to initialize: %v", err)
			return
		}
	} else if result, err = agent.InitLocal(); err != nil {
		warnf("required parameter missing, please run <zoomeye init -h> for help")
		return
	}
	successf("succeed to initialize")
	infof("ZoomEye Resources Info", "Role:  %s\nQuota: %d", result.Plan, result.Resources.Search)
}

func cmdInfo(agent *ZoomEyeAgent) {
	result, err := agent.Info()
	if err != nil {
		checkError(err)
		return
	}
	successf("succeed to query")
	infof("ZoomEye Resources Info", "Role:  %s\nQuota: %d", result.Plan, result.Resources.Search)
}

func cmdSearch(agent *ZoomEyeAgent) {
	var (
		analyzer = newResultAnalyzer()
		flgs     struct {
			num      int    `value:"20" usage:"The number of search results that should be returned, multiple of 20"`
			resource string `name:"type" usage:"Specify the type of resource to search"`
			force    bool   `usage:"Ignore local and cache data"`
		}
		args = parseFlags("search", &flgs, `"weblogic" -facet "app" -count`)
	)
	if len(args) == 0 {
		warnf("search keyword missing, please run <zoomeye search -h> for help")
		return
	}
	var (
		dork        = args[0]
		start       = time.Now()
		result, err = agent.Search(dork, flgs.num, flgs.resource, flgs.force)
		since       = time.Since(start)
	)
	if err != nil {
		checkError(err)
		return
	}
	successf("succeed to search (in %v)", since)
	analyzer.do(result, func(filtered []map[string]interface{}) {
		name := fmt.Sprintf("%s_%s_%d", flgs.resource, url.QueryEscape(dork), flgs.num)
		if path, err := agent.Save(name, result); err != nil {
			errorf("failed to save: %v", err)
		} else {
			successf("succeed to save (%s)", path)
			agent.SaveFilterData(filepath.Join(agent.conf.DataPath, name+"_filtered.json"), filtered)
		}
	})
}

func cmdLoad(agent *ZoomEyeAgent) {
	var (
		analyzer = newResultAnalyzer()
		args     = parseFlags("load", nil, `"data/host_weblogic_20.json" -facet "app" -count`)
	)
	if len(args) == 0 {
		warnf("path of local data file missing, please run <zoomeye load -h> for help")
		return
	}
	var (
		file        = args[0]
		result, err = agent.Load(file)
	)
	if err != nil {
		errorf("invalid local data: %v", err)
		return
	}
	successf("succeed to load")
	analyzer.do(result, func(filtered []map[string]interface{}) {
		var (
			ext  = filepath.Ext(file)
			path = strings.TrimSuffix(file, ext) + "_filtered" + ext
		)
		if err := agent.SaveFilterData(path, filtered); err != nil {
			errorf("failed to save: %v", err)
		} else {
			path, _ = filepath.Abs(path)
			successf("succeed to save (%s)", path)
		}
	})
}

func cmdHistory(agent *ZoomEyeAgent) {
	var (
		flgs struct {
			filter string `usage:"Output more clearer query results by set filter field"`
			num    int    `value:"20" usage:"The number of results that should be returned"`
			force  bool   `usage:"Ignore cache data"`
		}
		args = parseFlags("history", &flgs, `"0.0.0.0" -filter "time=^2020-03,port,service" -num 1`)
	)
	if len(args) == 0 {
		warnf("ip missing, please run <zoomeye history -h> for help")
		return
	}
	var (
		start       = time.Now()
		result, err = agent.History(args[0], flgs.force)
		since       = time.Since(start)
	)
	if err != nil {
		checkError(err)
		return
	}
	successf("succeed to query (in %v)", since)
	showHistory(result, strings.Split(flgs.filter, ","), flgs.num)
}

func cmdClear(agent *ZoomEyeAgent) {
	var flgs struct {
		cache   bool
		setting bool
	}
	parseFlags("clear", &flgs, `-cache`, `-cache -setting`)
	agent.Clear(flgs.cache, flgs.setting)
	successf("succeed to clear data")
}
