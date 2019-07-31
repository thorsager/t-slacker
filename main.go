package main

import (
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
)

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("unable to detect current user: %v", err)
	}
	appHome = path.Join(usr.HomeDir, "."+constants.Name)
}

func main() {

	ctx, err := runtime.New(appHome,
		fmt.Sprintf(" %s %s - %s", constants.Name, version, constants.Url))
	if err != nil {
		log.Fatalf("unable to create application runtime: %v", err)
	}

	err = ctx.Run()
	if err != nil {
		log.Fatalf("unable to start application runtime: %v", err)
	}

}
