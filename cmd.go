package main

import (
	"ZoomEye-go/zoomeye"
	"flag"
	"fmt"
	"net/url"
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
		warnf("input parameter error, please run <zoomeye init -h> for help")
		return
	}
	successf("succeed to initialize")
	infof("Resources Info", "Role:  %s\nQuota: %d", result.Plan, result.Resources.Search)
}

func cmdInfo(agent *ZoomEyeAgent) {
	result, err := agent.Info()
	if err != nil {
		if _, ok := err.(*zoomeye.ErrorResult); ok {
			errorf("failed to authenticate: %v", err)
		} else {
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
		facet    string
		force    bool
		save     bool
		count    bool
		filter   string
		stat     string
	}
	flag.StringVar(&args.dork, "dork", "", "The ZoomEye search keyword or ZoomEye exported file")
	flag.IntVar(&args.num, "num", 20, "The number of search results that should be returned, multiple of 20")
	flag.StringVar(&args.resource, "type", "host", "Specify the type of resource to search")
	flag.StringVar(&args.facet, "facet", "", "Perform statistics on ZoomEye database")
	flag.BoolVar(&args.force, "force", false, "Ignore local and cache data")
	flag.BoolVar(&args.save, "save", false, "Save the search results in JSON format")
	flag.BoolVar(&args.count, "count", false, "The total number of results in ZoomEye database for a search")
	flag.StringVar(&args.filter, "filter", "", "Output more clearer search results by set filter field")
	flag.StringVar(&args.stat, "stat", "", "Perform statistics on search results")
	if flag.Parse(); args.dork == "" {
		warnf("input parameter error, please run <zoomeye search -h> for help")
		return
	}
	var (
		start       = time.Now()
		result, err = agent.Search(args.dork, args.num, args.resource, args.force)
		since       = time.Since(start)
	)
	if err != nil {
		if _, ok := err.(*zoomeye.ErrorResult); ok {
			errorf("failed to authenticate: %v", err)
		} else {
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
	if args.facet == "" && args.stat == "" && args.filter == "" {
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
