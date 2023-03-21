package main

import (
	"fmt"

	"github.com/leemos-xx/basin/common"

	"github.com/pkg/errors"

	"github.com/leemos-xx/basin/container"
	"github.com/leemos-xx/basin/network"
	"github.com/urfave/cli"
)

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		err := container.RunContainerInitProcess()
		return err
	},
}

// eg: basin run -it -name base base-1.0.0 /bin/bash
var runCmd = cli.Command{
	Name:  "run",
	Usage: "Run a command in a new lightweight container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "Enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "Run container in background",
		},
		cli.StringFlag{
			Name:  "mem",
			Usage: "Memory limit",
		},
		cli.StringFlag{
			Name:  "cpu",
			Usage: "Limit cpu cfs quota",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "CPUs in which to allow execution",
		},
		cli.StringFlag{
			Name:  "volume",
			Usage: "Bind mount a volume",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "Set container name",
		},
		cli.StringSliceFlag{
			Name:  "env",
			Usage: "Set environment variables",
		},
		cli.StringFlag{
			Name:  "network",
			Usage: "Connect a container to a network",
		},
		cli.StringSliceFlag{
			Name:  "port",
			Usage: "Expose a port or a range of ports",
		},
	},
	Action: func(context *cli.Context) error {
		// 命令行参数预校验
		if len(context.Args()) < 2 {
			return errors.New("invalid parameters")
		}
		// tty&detach 不能同时出现
		tty := context.Bool("it")
		detach := context.Bool("d")
		if tty && detach {
			return errors.New("it and d parameter can not both provided")
		}

		params := &common.RunParam{
			TTY:               tty,
			ContainerName:     context.String("name"),
			Envs:              context.StringSlice("env"),
			Network:           context.String("network"),
			PortMapping:       context.StringSlice("port"),
			Volume:            context.String("volume"),
			ImageName:         context.Args()[0],
			ContainerCommands: context.Args()[1:],
			CgroupConfig: &common.CgroupParam{
				CpuCfsQuota: context.Int("cpu"),
				CpuSet:      context.String("cpuset"),
				MemoryLimit: context.String("mem"),
			},
		}

		container.Run(params)

		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		container.ListContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("please input your container name")
		}
		containerName := context.Args().Get(0)
		container.ReadContainerLog(containerName)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args().Get(0)
		container.Stop(containerName)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args().Get(0)
		container.Remove(containerName)
		return nil
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}
				err := network.Init()
				if err != nil {
					return fmt.Errorf("init network error: %+v", err)
				}
				err = network.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args()[0])
				if err != nil {
					return fmt.Errorf("create network error: %+v", err)
				}
				return nil
			},
		},
		{
			Name:  "ps",
			Usage: "list container network",
			Action: func(context *cli.Context) error {
				err := network.Init()
				if err != nil {
					return err
				}
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:  "rm",
			Usage: "remove container network",
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}
				network.Init()
				err := network.DeleteNetwork(context.Args()[0])
				if err != nil {
					return fmt.Errorf("remove network error: %+v", err)
				}
				return nil
			},
		},
	},
}
