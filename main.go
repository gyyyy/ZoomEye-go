package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func help() {
	fmt.Printf("Usage of %s:\n"+
		"  init\n        Initialize ZoomEye by username/password or API-Key\n"+
		"  info\n        Query resources information\n"+
		"  search\n        Search results from local, cache or API\n"+
		"  load\n        Load results from local data file\n"+
		"  clean\n        Removes all cache data\n"+
		"  help\n        Usage of ZoomEye-go\n",
		filepath.Base(os.Args[0]))
}

func main() {
	var (
		agent = NewAgent()
		n     = len(os.Args)
	)
	if n == 1 {
		// TODO: user interact mode
		warnf("User-Interact mode is coming soon, please run <zoomeye -h> for help")
		return
	}
	cmd := strings.ToLower(os.Args[1])
	if n > 2 {
		os.Args = append(os.Args[0:1], os.Args[2:]...)
	}
	switch cmd {
	case "init":
		cmdInit(agent)
	case "info":
		cmdInfo(agent)
	case "search":
		cmdSearch(agent)
	case "load":
		cmdLoad(agent)
	case "clean":
		cmdClean(agent)
	case "help", "-help", "--help", "-h", "?":
		help()
	default:
		warnf("input parameter error, please run <zoomeye -h> for help")
	}
}
