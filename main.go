package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ver = "v1.6"

func banner() {
	fmt.Println(
		colorf("\n         ,----,\n"+
			"       .'   .`|                            ____      ,---,.\n"+
			"    .'   .'   ;                          ,'  , `.  ,'  .' |\n"+
			"  ,---, '    .' ,---.     ,---.       ,-+-,.' _ |,---.'   |            ", colorLightYellow),
		colorf(ver, colorLightWhite),
		colorf("\n  |   :     ./ '   ,'\\   '   ,'\\   ,-+-. ;   , |||   |   .'\n"+
			"  ;   | .'  / /   /   | /   /   | ,--.'|'   |  ||:   :  |-,      .--,   ,---.\n"+
			"  `---' /  ; .   ; ,. :.   ; ,. :|   |  ,', |  |,:   |  ;/|    /_ ./|  /     \\\n"+
			"    /  ;  /  '   | |: :'   | |: :|   | /  | |--' |   :   .' , ' , ' : /    /  |\n"+
			"   ;  /  /--,'   | .; :'   | .; :|   : |  | ,    |   |  |-,/___/ \\: |.    ' / |\n"+
			"  /  /  / .`||   :    ||   :    ||   : |  |/     '   :  ;/| .  \\  ' |'   ;   /|\n"+
			"./__;       : \\   \\  /  \\   \\  / |   | |`-'      |   |    \\  \\  ;   :'   |  / |\n"+
			"|   :     .'   `----'    `----'  |   ;/          |   :   .'   \\  \\  ;|   :    |\n"+
			";   |  .'                        '---'           |   | ,'      :  \\  \\\\   \\  /\n"+
			"`---'  ", colorLightYellow),
		colorf("<https://github.com/gyyyy/ZoomEye-go/>", colorLightWhite),
		colorf("  `----'         \\  ' ; `----'\n"+
			"                                                                 `--`\n", colorLightYellow),
	)
}

func help() {
	fmt.Printf("Usage of %s:\n"+
		"  version\n        Show current version\n"+
		"  init\n        Initialize ZoomEye by username/password or API-Key\n"+
		"  info\n        Query resources information\n"+
		"  search\n        Search results from local, cache or API\n"+
		"  load\n        Load results from local data file\n"+
		"  history\n        Query device history\n"+
		"  clear\n        Removes all cache and setting data\n"+
		"  help\n        Usage of ZoomEye-go\n",
		filepath.Base(os.Args[0]))
}

func main() {
	var (
		agent = NewAgent()
		cmd   string
	)
	if len(os.Args) > 1 {
		cmd = os.Args[1]
		os.Args = append(os.Args[0:1], os.Args[2:]...)
	}
	switch strings.ToLower(cmd) {
	case "init":
		cmdInit(agent)
	case "info":
		cmdInfo(agent)
	case "search":
		cmdSearch(agent)
	case "load":
		cmdLoad(agent)
	case "history":
		cmdHistory(agent)
	case "clear":
		cmdClear(agent)
	case "version", "-version", "--version", "ver", "-ver", "--ver", "-v", "--v":
		banner()
	case "help", "-help", "--help", "-h", "--h", "?":
		help()
	case "":
		warnf("Cli-User-Interact mode is coming soon, please run <zoomeye -h> for help")
	default:
		warnf("unsupported command please run <zoomeye -h> for help")
	}
}
