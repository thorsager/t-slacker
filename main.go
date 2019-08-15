package main

import (
	"flag"
	"fmt"
	"github.com/thorsager/t-slacker/constants"
	"github.com/thorsager/t-slacker/runtime"
	"log"
	"os/user"
	"path"
)

var (
	version = "undefined"
	appHome string
	debug   bool
)

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("unable to detect current user: %v", err)
	}
	appHome = path.Join(usr.HomeDir, "."+constants.Name)
	flag.BoolVar(&debug, "debug", false, "Enable debug")
	flag.Parse()
}

func main() {
	log.Printf("deug=%t", debug)
	ctx, err := runtime.New(appHome,
		fmt.Sprintf(" %s %s - %s", constants.Name, version, constants.Url),
		debug)
	if err != nil {
		log.Fatalf("unable to create application runtime: %v", err)
	}

	err = ctx.Run()
	if err != nil {
		log.Fatalf("unable to start application runtime: %v", err)
	}

}
