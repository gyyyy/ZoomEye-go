package main

import (
	"ZoomEye-go/zoomeye"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func usage(cmd string, examples ...string) func() {
	var example string
	if len(examples) > 0 {
		example = "Example:\n"
		for _, v := range examples {
			example += fmt.Sprintf("  ./ZoomEye-go %s %s\n", cmd, v)
		}
		example += "\n"
	}
	return func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s (%s):\n", filepath.Base(os.Args[0]), cmd)
		flag.PrintDefaults()
		if example != "" {
			fmt.Fprintf(flag.CommandLine.Output(), example)
		}
	}
}

func cmdInit(agent *ZoomEyeAgent) {
	var args struct {
		apiKey   string
		username string
		password string
	}
	flag.StringVar(&args.apiKey, "apikey", "", "ZoomEye API-Key")
	flag.StringVar(&args.username, "username", "", "ZoomEye account username")
	flag.StringVar(&args.password, "password", "", "ZoomEye account password")
	flag.Usage = usage("init", `-apikey "XXXXXXXX-XXXX-XXXXX-XXXX-XXXXXXXXXXX"`,
		`-username "username@zoomeye.org" -password "password"`)
	flag.Parse()
	var (
		result *zoomeye.ResourcesInfoResult
		err    error
	)
	if args.apiKey != "" {
		if result, err = agent.InitByKey(args.apiKey); err != nil {
			errorf("failed to initialize: %v", err)
			return
		}
	} else if args.username != "" && args.password != "" {
		if result, err = agent.InitByUser(args.username, args.password); err != nil {
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

func cmdInfo(agent *ZoomEyeAgent) {
	result, err := agent.Info()
	if err != nil {
		checkError(err)
		return
	}
	successf("succeed to query")
	infof("ZoomEye Resources Info", "Role:  %s\nQuota: %d", result.Plan, result.Resources.Search)
}

func commandVal() (string, bool) {
	if len(os.Args) <= 1 {
		return "", false
	}
	v := os.Args[1]
	if strings.HasPrefix(v, "-") {
		return "", true
	}
	os.Args = append(os.Args[0:1], os.Args[2:]...)
	return v, true
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

func (a *resultAnalyzer) do(result *zoomeye.SearchResult, saveCallback func()) {
	if a.count {
		infof("ZoomEye Total", "Count: %d", result.Total)
	}
	if a.figure != "" {
		if a.figure = strings.ToLower(a.figure); a.figure != "pie" {
			a.figure = "hist"
		}
	}
	if a.facet != "" {
		printFacet(result, strings.Split(a.facet, ","), a.figure)
	}
	if a.stat != "" {
		printStat(result, strings.Split(a.stat, ","), a.figure)
	}
	if a.filter != "" {
		printFilter(result, strings.Split(a.filter, ","))
	}
	if !a.count && a.facet == "" && a.stat == "" && a.filter == "" {
		printData(result)
	}
	if a.save && saveCallback != nil {
		saveCallback()
	}
}

func cmdSearch(agent *ZoomEyeAgent) {
	dork, ok := commandVal()
	if !ok {
		warnf("search keyword missing, please run <zoomeye search -h> for help")
		return
	}
	var (
		analyzer = newResultAnalyzer()
		args     struct {
			num      int
			resource string
			force    bool
		}
	)
	flag.IntVar(&args.num, "num", 20, "The number of search results that should be returned, multiple of 20")
	flag.StringVar(&args.resource, "type", "host", "Specify the type of resource to search")
	flag.BoolVar(&args.force, "force", false, "Ignore local and cache data")
	flag.Usage = usage("search", `"weblogic" -facet "app" -count`)
	flag.Parse()
	var (
		start       = time.Now()
		result, err = agent.Search(dork, args.num, args.resource, args.force)
		since       = time.Since(start)
	)
	if err != nil {
		checkError(err)
		return
	}
	successf("succeed to search (in %v)", since)
	analyzer.do(result, func() {
		name := fmt.Sprintf("%s_%s_%d", args.resource, url.QueryEscape(dork), args.num)
		if path, err := agent.Save(name, result); err != nil {
			errorf("failed to save: %v", err)
		} else {
			successf("succeed to save (%s)", path)
		}
	})
}

func cmdLoad(agent *ZoomEyeAgent) {
	file, ok := commandVal()
	if !ok {
		warnf("path of local data file missing, please run <zoomeye load -h> for help")
		return
	}
	analyzer := newResultAnalyzer()
	flag.Usage = usage("load", `"data/host_weblogic_20.json" -facet "app" -count`)
	flag.Parse()
	result, err := agent.Load(file)
	if err != nil {
		errorf("invalid local data: %v", err)
		return
	}
	successf("succeed to load")
	analyzer.do(result, func() {
		var (
			ext  = filepath.Ext(file)
			path = strings.TrimSuffix(file, ext) + "_filtered" + ext
		)
		if err := agent.SaveFilterData(path, result.FilterCache); err != nil {
			errorf("failed to save: %v", err)
		} else {
			path, _ = filepath.Abs(path)
			successf("succeed to save (%s)", path)
		}
	})
}

func cmdClean(agent *ZoomEyeAgent) {
	agent.Clean()
	successf("succeed to clean all cache data")
}
