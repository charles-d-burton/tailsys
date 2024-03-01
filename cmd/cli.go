package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/charles-d-burton/tailsys/services/client"
	"github.com/charles-d-burton/tailsys/services/coordination"
	"github.com/urfave/cli/v2"
)

// Enum type for Auth Type
type AuthType int

// Enum definition for Auth Type
const (
	OAUTH AuthType = iota
	AUTHKEY
)

const (
	clientIdFlag    = "client-id"
	clienSecretFlag = "client-secret"
	authKeyFlag     = "auth-key"
	portFlag        = "port"
	hostnameFlag    = "hostname"
	tsnetVerbose    = "tsnet-verbose"
)

func StartCLI() error {
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
			coordinationServerCommand(),
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
		&cli.StringFlag{
			Name:        portFlag,
			Usage:       "port for rpc server to listen on, default 6655",
			DefaultText: "6655",
			EnvVars:     []string{"RPC_PORT", "PORT"},
		},
		&cli.StringFlag{
			Name:    hostnameFlag,
			Usage:   "set tailnet hostname",
			EnvVars: []string{"TS_HOSTNAME", "HOSTNAME"},
		},
		//TODO: This does not currently do anything
		&cli.BoolFlag{
			Name:  tsnetVerbose,
			Usage: "enable verbose tsnet logging",
		},
	}
}

func coordinationServerCommand() *cli.Command {
	return &cli.Command{
		Name:    "coordination-server",
		Aliases: []string{"co"},
		Usage:   "Start the application in server mode",
		Action: func(ctx *cli.Context) error {
			fmt.Println("starting the server code")
			var co coordination.Coordinator
			err := co.NewCoordinator(ctx.Context, co.WithDevMode(ctx.Bool("dev")))
			authType, err := getAuthType(ctx)
			if err != nil {
				return err
			}

			hostname := ctx.Value(hostnameFlag).(string)
			switch authType {
			case OAUTH:
				id := ctx.Value(clientIdFlag).(string)
				secret := ctx.Value(clienSecretFlag).(string)
				if err := co.ConnectOauth(ctx.Context, id, secret, hostname); err != nil {
					return err
				}
			case AUTHKEY:
				authkey := ctx.Value(authKeyFlag).(string)
				if err := co.ConnectAuthKey(ctx.Context, authkey, hostname); err != nil {
					return err
				}
			}
			return co.StartRPCCoordinationServer(ctx.Context)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dev",
				Usage: "set to enable dev mode, all node registration automatically accepted",
			},
		},
	}
}

func clientCommand() *cli.Command {
	return &cli.Command{
		Name:    "client",
		Aliases: []string{"c"},
		Usage:   "Start the application in client mode",
		Action: func(ctx *cli.Context) error {
			fmt.Println("starting the client code")

			cl, err := client.NewClient(ctx.Context)
			authType, err := getAuthType(ctx)
			if err != nil {
				return err
			}

			hostname := ctx.Value(hostnameFlag).(string)
			coServer := ctx.Value("coordination-server").(string)

			switch authType {
			case OAUTH:
				id := ctx.Value(clientIdFlag).(string)
				secret := ctx.Value(clienSecretFlag).(string)
				if err := cl.ConnectOauth(ctx.Context, id, secret, hostname); err != nil {
					return err
				}
			case AUTHKEY:
				authkey := ctx.Value(authKeyFlag).(string)
				if err := cl.ConnectAuthKey(ctx.Context, authkey, hostname); err != nil {
					return err
				}
			}
			if err := cl.RegisterWithCoordinationServer(ctx.Context, coServer); err != nil {
				return err
			}
			return cl.StartRPCClientMode(ctx.Context)

		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "tags",
				Usage: "tags for discovering coordination server on tailscale",
			},
			&cli.StringFlag{
				Name:    "coordination-server",
				Aliases: []string{"cs"},
				Usage:   "tags for discovering coordination server on tailscale",
				EnvVars: []string{"COORDINATION_SERVER", "C_SERVER"},
			},
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

func getAuthType(ctx *cli.Context) (AuthType, error) {
	id := ctx.Value(clientIdFlag).(string)
	secret := ctx.Value(clienSecretFlag).(string)
	if id != "" && secret != "" {
		return OAUTH, nil
	}

	authKey := ctx.Value(authKeyFlag).(string)
	if authKey != "" {
		return AUTHKEY, nil

	}
	return -1, errors.New("unable to start, no auth provided")
}
