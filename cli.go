package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func startCLI(ctx context.Context) error {
  app := &cli.App {
    Description: "A systems management application that rides the tailscale network",
    Flags: globalFlags(),
    Before: func(ctx *cli.Context) error {
      return nil
    },
    Commands: []*cli.Command {
      serverCommand(),
      clientCommand(),
      interactiveCommand(),
      nonInteractiveCommand(),
    },
  }

  if err := app.Run(os.Args); err != nil {
    return err
  }
  return nil
}

func globalFlags() []cli.Flag{
  return []cli.Flag{
      &cli.StringFlag{
        Name: "client-id",
        Usage: "oauth clientid from tailscale",
        EnvVars: []string{"TS_CLIENT_ID", "CLIENT_ID"},
      },
      &cli.StringFlag{
        Name: "client-secret",
        Usage: "oauth clientsecret from tailscale",
        EnvVars: []string{"TS_CLIENT_SECRET", "CLIENT_SECRET"},

      },
      &cli.StringFlag{
        Name: "auth-key",
        Usage: "pre-generated auth key from tailcale",
        EnvVars: []string{"TS_AUTH_KEY", "AUTH_KEY"},
      },
  }
}

func serverCommand() *cli.Command {
  return &cli.Command{
    Name: "server",
    Aliases: []string{"s"},
    Usage: "Start the application in server mode",
    Action: func(ctx *cli.Context) error {
      fmt.Println("starting the server code")
      if ctx.NArg() == 0 {
        return errors.New("no tailcale auth supplied..")
      }

      return nil
    },
  }
}

func clientCommand() *cli.Command {
  return &cli.Command{
    Name: "client",
    Aliases: []string{"c"},
    Usage: "Start the application in client mode",
    Action: func(*cli.Context) error {
      fmt.Println("starting the client code")
      return nil
    },
  }
}

func interactiveCommand() *cli.Command {
  return &cli.Command{
    Name: "interactive",
    Aliases: []string{"i"},
    Usage: "Start the application interactive ui",
    Action: func(*cli.Context) error {
      fmt.Println("starting the interactive code")
      return nil
    },
  }
}

func nonInteractiveCommand() *cli.Command {
  return &cli.Command{
    Name: "non-interactive",
    Aliases: []string{"ni"},
    Usage: "Start the application non-interactively",
    Subcommands: []*cli.Command {
      {
        Name: "command",
        Usage: "shell command to run on remote",
        Action: func(cCtx *cli.Context) error {
          fmt.Println("running command runner")
          return nil
        },
        Subcommands: []*cli.Command{
          {
            Name: "machine-pattern",
            Aliases: []string{"mp"},
            Usage: "shell command to run on remote",
            Action: func(cCtx *cli.Context) error {
              fmt.Println("running command runner")
              return nil
            },
          },
        },
      },
      {
        Name: "file",
        Usage: "load a configuration file to run",
        Action: func(cCtx *cli.Context) error {
          fmt.Println("loading a configuration file")
          return nil
        },
      },
      {
        Name: "dir",
        Usage: "load a directory of configuration files to run",
        Action: func(cCtx *cli.Context) error {
          fmt.Println("loading a configuration files from dir")
          return nil
        },
      },
    },
  }
}

