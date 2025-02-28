package main

import (
    "os"

    log "github.com/sirupsen/logrus"
    "github.com/urfave/cli/v3"
)

const usage = `mydocker is a simple container runtime implementation.`

func main() {
    app := &cli.App{
        Name:  "mydocker",
        Usage: usage,
        Commands: []*cli.Command{
            initCommand,
            runCommand,
            listCommand,
            logCommand,
            execCommand,
            stopCommand,
            removeCommand,
            commitCommand,
            networkCommand,
        },
        Before: func(ctx *cli.Context) error {
            log.SetFormatter(&log.JSONFormatter{})
            log.SetOutput(os.Stdout)
            return nil
        },
    }

    if err := app.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}

// 示例命令定义
var runCommand = &cli.Command{
    Name:  "run",
    Usage: "Run a container",
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:    "mem",
            Usage:   "Memory limit",
            Aliases: []string{"m"},
        },
    },
    Action: func(ctx *cli.Context) error {
        mem := ctx.String("mem")
        log.Infof("Running with memory limit: %s", mem)
        return nil
    },
}
