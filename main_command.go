package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"github.com/yarqwq/mydocker/cgroups/subsystems"
	"github.com/yarqwq/mydocker/container"
	"github.com/yarqwq/mydocker/network"
	"os"
)

var runCommand = &cli.Command{
	Name:  "run",
	Usage: `Create a container with namespace and cgroups limit ie: mydocker run -ti [image] [command]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "ti",
			Usage:   "enable tty",
			Aliases: []string{"t"},
		},
		&cli.BoolFlag{
			Name:    "d",
			Usage:   "detach container",
			Aliases: []string{"detach"},
		},
		&cli.StringFlag{
			Name:    "m",
			Usage:   "memory limit",
			Aliases: []string{"memory"},
		},
		&cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		&cli.StringFlag{
			Name:    "name",
			Usage:   "container name",
			Aliases: []string{"n"},
		},
		&cli.StringFlag{
			Name:    "v",
			Usage:   "volume",
			Aliases: []string{"volume"},
		},
		&cli.StringSliceFlag{
			Name:    "e",
			Usage:   "set environment",
			Aliases: []string{"env"},
		},
		&cli.StringFlag{
			Name:    "net",
			Usage:   "container network",
			Aliases: []string{"network"},
		},
		&cli.StringSliceFlag{
			Name:    "p",
			Usage:   "port mapping",
			Aliases: []string{"publish"},
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 1 {
			return fmt.Errorf("missing container command")
		}
		cmdArray := ctx.Args().Slice()

		// 获取镜像名称
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]

		createTty := ctx.Bool("ti")
		detach := ctx.Bool("d")

		if createTty && detach {
			return fmt.Errorf("ti and d parameters cannot both be provided")
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: ctx.String("m"),
			CpuSet:      ctx.String("cpuset"),
			CpuShare:    ctx.String("cpushare"),
		}

		containerName := ctx.String("name")
		volume := ctx.String("v")
		network := ctx.String("net")
		envSlice := ctx.StringSlice("e")
		portmapping := ctx.StringSlice("p")

		log.Infof("Creating container with tty: %v", createTty)
		Run(createTty, cmdArray, resConf, containerName, volume, imageName, envSlice, network, portmapping)
		return nil
	},
}

var initCommand = &cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(ctx *cli.Context) error {
		log.Info("Initializing container...")
		return container.RunContainerInitProcess()
	},
}

var listCommand = &cli.Command{
	Name:  "ps",
	Usage: "List all containers",
	Action: func(ctx *cli.Context) error {
		ListContainers()
		return nil
	},
}

var logCommand = &cli.Command{
	Name:  "logs",
	Usage: "Print logs of a container",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 1 {
			return fmt.Errorf("please provide container name")
		}
		logContainer(ctx.Args().First())
		return nil
	},
}

var execCommand = &cli.Command{
	Name:  "exec",
	Usage: "Execute command in container",
	Action: func(ctx *cli.Context) error {
		if os.Getenv(ENV_EXEC_PID) != "" {
			log.Infof("PID callback: %d", os.Getpid())
			return nil
		}

		if ctx.Args().Len() < 2 {
			return fmt.Errorf("missing container name or command")
		}

		containerName := ctx.Args().First()
		commandArray := ctx.Args().Tail()
		ExecContainer(containerName, commandArray)
		return nil
	},
}

var stopCommand = &cli.Command{
	Name:  "stop",
	Usage: "Stop a container",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().IsEmpty() {
			return fmt.Errorf("missing container name")
		}
		stopContainer(ctx.Args().First())
		return nil
	},
}

var removeCommand = &cli.Command{
	Name:  "rm",
	Usage: "Remove a container",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().IsEmpty() {
			return fmt.Errorf("missing container name")
		}
		removeContainer(ctx.Args().First())
		return nil
	},
}

var commitCommand = &cli.Command{
	Name:  "commit",
	Usage: "Commit container to image",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 2 {
			return fmt.Errorf("missing container name and image name")
		}
		commitContainer(ctx.Args().Get(0), ctx.Args().Get(1))
		return nil
	},
}

var networkCommand = &cli.Command{
	Name:  "network",
	Usage: "Manage container networks",
	Subcommands: []*cli.Command{
		{
			Name:  "create",
			Usage: "Create a network",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "driver",
					Usage: "Network driver type",
				},
				&cli.StringFlag{
					Name:  "subnet",
					Usage: "Subnet CIDR",
				},
			},
			Action: func(ctx *cli.Context) error {
				if ctx.Args().IsEmpty() {
					return fmt.Errorf("missing network name")
				}
				network.Init()
				return network.CreateNetwork(
					ctx.String("driver"),
					ctx.String("subnet"),
					ctx.Args().First(),
				)
			},
		},
		{
			Name:  "list",
			Usage: "List networks",
			Action: func(ctx *cli.Context) error {
				network.Init()
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "Remove network",
			Action: func(ctx *cli.Context) error {
				if ctx.Args().IsEmpty() {
					return fmt.Errorf("missing network name")
				}
				network.Init()
				return network.DeleteNetwork(ctx.Args().First())
			},
		},
	},
}

func main() {
	app := &cli.App{
		Name:  "mydocker",
		Usage: "A simple container runtime implementation",
		Commands: []*cli.Command{
			runCommand,
			initCommand,
			listCommand,
			logCommand,
			execCommand,
			stopCommand,
			removeCommand,
			commitCommand,
			networkCommand,
		},
		Before: func(ctx *cli.Context) error {
			// 初始化日志配置
			log.SetFormatter(&log.JSONFormatter{})
			log.SetOutput(os.Stdout)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
