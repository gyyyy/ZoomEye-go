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

func cmdInit(agent *ZoomEyeAgent) {
	var args struct {
		apiKey   string
		username string
		password string
	}
	flag.StringVar(&args.apiKey, "apikey", "", "ZoomEye API-Key")
	flag.StringVar(&args.username, "username", "", "ZoomEye account username")
	flag.StringVar(&args.password, "password", "", "ZoomEye account password")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s (init):\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(),
			"Example:\n  ./ZoomEye-go init -apikey \"XXXXXXXX-XXXX-XXXXX-XXXX-XXXXXXXXXXX\"\n"+
				"  ./ZoomEye-go init -username \"username@zoomeye.org\" -password \"password\"\n\n")
	}
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

func cmdInfo(agent *ZoomEyeAgent) {
	result, err := agent.Info()
	if err != nil {
		switch err.(type) {
		case *NoAuthKeyErr:
			warnf("not found any Auth Keys, please run <zoomeye init> first")
		case *zoomeye.ErrorResult:
			errorf("failed to authenticate: %v", err)
		default:
			errorf("something is wrong: %v", err)
		}
		return
	}
	successf("succeed to query")
	infof("ZoomEye Resources Info", "Role:  %s\nQuota: %d", result.Plan, result.Resources.Search)
}

func commandVal() string {
	var v string
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		v = os.Args[1]
		os.Args = append(os.Args[0:1], os.Args[2:]...)
	}
	return v
}

func cmdSearch(agent *ZoomEyeAgent) {
	dork := commandVal()
	if dork == "" {
		warnf("search keyword missing, please run <zoomeye search -h> for help")
		return
	}
	var args struct {
		num      int
		resource string
		force    bool
		count    bool
		facet    string
		stat     string
		figure   string
		filter   string
		save     bool
	}
	flag.IntVar(&args.num, "num", 20, "The number of search results that should be returned, multiple of 20")
	flag.StringVar(&args.resource, "type", "host", "Specify the type of resource to search")
	flag.BoolVar(&args.force, "force", false, "Ignore local and cache data")
	flag.BoolVar(&args.count, "count", false, "The total number of results in ZoomEye database for a search")
	flag.StringVar(&args.facet, "facet", "", "Perform statistics on ZoomEye database")
	flag.StringVar(&args.stat, "stat", "", "Perform statistics on search results")
	flag.StringVar(&args.figure, "figure", "", "Output Pie or bar chart only be used under facet and stat")
	flag.StringVar(&args.filter, "filter", "", "Output more clearer search results by set filter field")
	flag.BoolVar(&args.save, "save", false, "Save the search results in JSON format")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s (search):\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(),
			"Example:\n  ./Zoomeye-go search \"weblogic\" -facet \"app\" -count\n\n")
	}
	flag.Parse()
	var (
		start       = time.Now()
		result, err = agent.Search(dork, args.num, args.resource, args.force)
		since       = time.Since(start)
	)
	if err != nil {
		switch err.(type) {
		case *NoAuthKeyErr:
			warnf("not found any Auth Keys, please run <zoomeye init> first")
		case *zoomeye.ErrorResult:
			errorf("failed to authenticate: %v", err)
		default:
			errorf("something is wrong: %v", err)
		}
		return
	}
	successf("succeed to search (in %v)", since)
	if args.count {
		infof("ZoomEye Total", "Count: %d", result.Total)
	}
	if args.figure != "" {
		if args.figure = strings.ToLower(args.figure); args.figure != "pie" {
			args.figure = "hist"
		}
	}
	if args.facet != "" {
		printFacet(result, strings.Split(args.facet, ","), args.figure)
	}
	if args.stat != "" {
		printStat(result, strings.Split(args.stat, ","), args.figure)
	}
	if args.filter != "" {
		printFilter(result, strings.Split(args.filter, ","))
	}
	if !args.count && args.facet == "" && args.stat == "" && args.filter == "" {
		printData(result)
	}
	if args.save {
		name := fmt.Sprintf("%s_%s_%d", args.resource, url.QueryEscape(dork), args.num)
		if path, err := agent.Save(result, name); err != nil {
			errorf("failed to save: %v", err)
		} else {
			successf("succeed to save (%s)", path)
		}
	}
}

func cmdLoad(agent *ZoomEyeAgent) {
	file := commandVal()
	if file == "" {
		warnf("path of local data file missing, please run <zoomeye load -h> for help")
		return
	}
	var args struct {
		count  bool
		facet  string
		stat   string
		figure string
		filter string
		save   bool
	}
	flag.BoolVar(&args.count, "count", false, "The total number of results in ZoomEye database in local data")
	flag.StringVar(&args.facet, "facet", "", "Perform statistics on ZoomEye database in local data")
	flag.StringVar(&args.stat, "stat", "", "Perform statistics in local data")
	flag.StringVar(&args.figure, "figure", "", "Output Pie or bar chart only be used under facet and stat")
	flag.StringVar(&args.filter, "filter", "", "Output more clearer results by set filter field in local data")
	flag.BoolVar(&args.save, "save", false, "Save the filter data in JSON format")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s (load):\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(),
			"Example:\n  ./Zoomeye-go load \"data/host_weblogic_20.json\" -facet \"app\" -count\n\n")
	}
	flag.Parse()
	result, err := agent.Load(file)
	if err != nil {
		errorf("invalid local data: %v", err)
		return
	}
	successf("succeed to load")
	if args.count {
		infof("ZoomEye Total", "Count: %d", result.Total)
	}
	if args.figure != "" {
		if args.figure = strings.ToLower(args.figure); args.figure != "pie" {
			args.figure = "hist"
		}
	}
	if args.facet != "" {
		printFacet(result, strings.Split(args.facet, ","), args.figure)
	}
	if args.stat != "" {
		printStat(result, strings.Split(args.stat, ","), args.figure)
	}
	if args.filter != "" {
		printFilter(result, strings.Split(args.filter, ","))
	}
	if !args.count && args.facet == "" && args.stat == "" && args.filter == "" {
		printData(result)
	}
	if args.save {
		var (
			ext  = filepath.Ext(file)
			path = strings.TrimSuffix(file, ext) + "_filtered" + ext
		)
		if err := agent.SaveFilterData(result.FilterCache, path); err != nil {
			errorf("failed to save: %v", err)
		} else {
			path, _ = filepath.Abs(path)
			successf("succeed to save (%s)", path)
		}
	}
}

func cmdClean(agent *ZoomEyeAgent) {
	agent.Clean()
	successf("succeed to clean all cache data")
}
