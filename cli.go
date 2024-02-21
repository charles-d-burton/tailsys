package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const (
  clientIdFlag    = "client-id"
	clienSecretFlag = "client-secret"
	authKeyFlag     = "auth-key"
)

func startCLI(ctx context.Context) error {
	app := &cli.App{
		Name:        "tailsys",
		Description: "A systems management application that rides the tailscale network",
		Flags:       globalFlags(),
		Before: func(ctx *cli.Context) error {
			id := ctx.Value(clientIdFlag).(string)
			secret := ctx.Value(clienSecretFlag).(string)
			if id != "" && secret != "" {
				return nil
			}

			authKey := ctx.Value(authKeyFlag).(string)
			if authKey != "" {
				return nil
			}
			cli.ShowAppHelp(ctx)
			return errors.New("error, must set either the oauth client/secret or pass a pre-generated auth-key")
		},
		Commands: []*cli.Command{
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

func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    clientIdFlag,
			Usage:   "oauth clientid from tailscale",
			EnvVars: []string{"TS_CLIENT_ID", "CLIENT_ID"},
		},
		&cli.StringFlag{
			Name:    clienSecretFlag,
			Usage:   "oauth clientsecret from tailscale",
			EnvVars: []string{"TS_CLIENT_SECRET", "CLIENT_SECRET"},
		},
		&cli.StringFlag{
			Name:    authKeyFlag,
			Usage:   "pre-generated auth key from tailcale",
			EnvVars: []string{"TS_AUTH_KEY", "AUTH_KEY"},
		},
	}
}

func serverCommand() *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"s"},
		Usage:   "Start the application in server mode",
		Action: func(ctx *cli.Context) error {
			fmt.Println("starting the server code")
      return startGRPCConnection(ctx) 
		},
	}
}

func clientCommand() *cli.Command {
	return &cli.Command{
		Name:    "client",
		Aliases: []string{"c"},
		Usage:   "Start the application in client mode",
		Action: func(*cli.Context) error {
			fmt.Println("starting the client code")
			return nil
		},
	}
}

func interactiveCommand() *cli.Command {
	return &cli.Command{
		Name:    "interactive",
		Aliases: []string{"i"},
		Usage:   "Start the application interactive ui",
		Action: func(*cli.Context) error {
			fmt.Println("starting the interactive code")
			return nil
		},
	}
}

func nonInteractiveCommand() *cli.Command {
	return &cli.Command{
		Name:    "non-interactive",
		Aliases: []string{"ni"},
		Usage:   "Start the application non-interactively",
		Subcommands: []*cli.Command{
			{
				Name:  "command",
				Usage: "shell command to run on remote",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("running command runner")
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:    "machine-pattern",
						Aliases: []string{"mp"},
						Usage:   "shell command to run on remote",
						Action: func(cCtx *cli.Context) error {
							fmt.Println("running command runner")
							return nil
						},
					},
				},
			},
			{
				Name:  "file",
				Usage: "load a configuration file to run",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("loading a configuration file")
					return nil
				},
			},
			{
				Name:  "dir",
				Usage: "load a directory of configuration files to run",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("loading a configuration files from dir")
					return nil
				},
			},
		},
	}
}

func startGRPCConnection(ctx *cli.Context) error {
	id := ctx.Value(clientIdFlag).(string)
	secret := ctx.Value(clienSecretFlag).(string)
	if id != "" && secret != "" {
		_, err := connectOauth(ctx.Context, id, secret)
		if err != nil {
			return err
		}
		return nil
	}

	authKey := ctx.Value(authKeyFlag).(string)
	if authKey != "" {
		_, err := connectAuthKey(ctx.Context, authKey)
		if err != nil {
			return nil
		}
		return nil
	}
	return nil
}
