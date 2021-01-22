package main

import (
	"ZoomEye-go/zoomeye"
	"flag"
	"fmt"
	"net/url"
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
	infof("Resources Info", "Role:  %s\nQuota: %d", result.Plan, result.Resources.Search)
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
	infof("Resources Info", "Role:  %s\nQuota: %d", result.Plan, result.Resources.Search)
}

func cmdSearch(agent *ZoomEyeAgent) {
	var args struct {
		dork     string
		num      int
		resource string
		force    bool
		count    bool
		facet    string
		stat     string
		filter   string
		save     bool
	}
	flag.StringVar(&args.dork, "dork", "", "[REQUIRED] The ZoomEye search keyword or ZoomEye exported file")
	flag.IntVar(&args.num, "num", 20, "The number of search results that should be returned, multiple of 20")
	flag.StringVar(&args.resource, "type", "host", "Specify the type of resource to search")
	flag.BoolVar(&args.force, "force", false, "Ignore local and cache data")
	flag.BoolVar(&args.count, "count", false, "The total number of results in ZoomEye database for a search")
	flag.StringVar(&args.facet, "facet", "", "Perform statistics on ZoomEye database")
	flag.StringVar(&args.stat, "stat", "", "Perform statistics on search results")
	flag.StringVar(&args.filter, "filter", "", "Output more clearer search results by set filter field")
	flag.BoolVar(&args.save, "save", false, "Save the search results in JSON format")
	if flag.Parse(); args.dork == "" {
		warnf("required parameter missing, please run <zoomeye search -h> for help")
		return
	}
	var (
		start       = time.Now()
		result, err = agent.Search(args.dork, args.num, args.resource, args.force)
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
		infof("Total Count", "Count: %d", result.Total)
	}
	if args.facet != "" {
		printFacet(result, strings.Split(args.facet, ","))
	}
	if args.stat != "" {
		printStat(result, strings.Split(args.stat, ","))
	}
	if args.filter != "" {
		printFilter(result, strings.Split(args.filter, ","))
	}
	if !args.count && args.facet == "" && args.stat == "" && args.filter == "" {
		printData(result)
	}
	if args.save {
		name := fmt.Sprintf("%s_%s_%d", args.resource, url.QueryEscape(args.dork), args.num)
		if path, err := agent.Save(result, name); err != nil {
			errorf("failed to save: %v", err)
		} else {
			successf("succeed to save (%s)", path)
		}
	}
}

func cmdLoad(agent *ZoomEyeAgent) {
	var args struct {
		file   string
		count  bool
		facet  string
		stat   string
		filter string
		save   bool
	}
	flag.StringVar(&args.file, "file", "", "[REQUIRED] The path of local data")
	flag.BoolVar(&args.count, "count", false, "The total number of results in ZoomEye database in local data")
	flag.StringVar(&args.facet, "facet", "", "Perform statistics on ZoomEye database in local data")
	flag.StringVar(&args.stat, "stat", "", "Perform statistics in local data")
	flag.StringVar(&args.filter, "filter", "", "Output more clearer results by set filter field in local data")
	flag.BoolVar(&args.save, "save", false, "Save the filter data in JSON format")
	if flag.Parse(); args.file == "" {
		warnf("required parameter missing, please run <zoomeye load -h> for help")
		return
	}
	result, err := agent.Load(args.file)
	if err != nil {
		errorf("invalid local data: %v", err)
		return
	}
	successf("succeed to load")
	if args.count {
		infof("Total Count", "Count: %d", result.Total)
	}
	if args.facet != "" {
		printFacet(result, strings.Split(args.facet, ","))
	}
	if args.stat != "" {
		printStat(result, strings.Split(args.stat, ","))
	}
	if args.filter != "" {
		printFilter(result, strings.Split(args.filter, ","))
	}
	if !args.count && args.facet == "" && args.stat == "" && args.filter == "" {
		printData(result)
	}
	if args.save {
		var (
			ext  = filepath.Ext(args.file)
			path = strings.TrimSuffix(args.file, ext) + "_filtered" + ext
		)
		if err := agent.SaveFilterData(result.FilterCache, path); err != nil {
			errorf("failed to save: %v", err)
		} else {
			path, _ = filepath.Abs(path)
			successf("succeed to save (%s)", path)
		}
	}
}
