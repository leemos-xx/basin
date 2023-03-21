package main

import (
	"log"
	"os"

	"github.com/leemos-xx/basin/common"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = common.Basin

	app.Commands = []cli.Command{
		initCommand,
		runCmd,
		listCommand,
		logCommand,
		stopCommand,
		removeCommand,
		networkCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
