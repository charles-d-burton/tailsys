package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func startCLI() error {
  app := &cli.App {
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name: "client-id",
        Usage: "oauth clientid from tailscale",
        EnvVars: []string{"TS_CLIENT_ID", "CLIENT_ID"},
      },
      &cli.StringFlag{
        Name: "client-id",
        Usage: "oauth clientsecret from tailscale",
        EnvVars: []string{"TS_CLIENT_SECRET", "CLIENT_SECRET"},

      },
      &cli.StringFlag{
        Name: "auth-key",
        Usage: "pre-generated auth key from tailcale",
        EnvVars: []string{"TS_AUTH_KEY", "AUTH_KEY"},
      },
    },
    Commands: []*cli.Command {
      {
        Name: "server",
        Aliases: []string{"s"},
        Usage: "Start the application in server mode",
        Action: func(*cli.Context) error {
          fmt.Println("starting the server code")
          return nil
        },
      },
      {
        Name: "client",
        Aliases: []string{"c"},
        Usage: "Start the application in client mode",
        Action: func(*cli.Context) error {
          fmt.Println("starting the client code")
          return nil
        },
      },
      {
        Name: "interactive",
        Aliases: []string{"i"},
        Usage: "Start the application interactive ui",
        Action: func(*cli.Context) error {
          fmt.Println("starting the interactive code")
          return nil
        },
      },
      {
        Name: "non-interactive",
        Aliases: []string{"ni"},
        Usage: "Start the application non-interactively",
        Subcommands: []*cli.Command {
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
      },
    },
  }

  if err := app.Run(os.Args); err != nil {
    return err
  }
  return nil
}
