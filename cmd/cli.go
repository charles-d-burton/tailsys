package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charles-d-burton/tailsys/services/client"
	"github.com/charles-d-burton/tailsys/services/commander"
	"github.com/charles-d-burton/tailsys/services/coordination"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)
func initConfig(cmd *cobra.Command) error {
	v := viper.New()
	if Check() {
		v.SetConfigFile("/var/lib/tailsys/config")
	} else {
		//find the home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		v.AddConfigPath(filepath.Join(home, ".local", "tailsys"))
		v.SetConfigType("yaml")
		v.SetConfigName("tailsys")
	}

	if err := v.ReadInConfig(); err == nil {
		//It's ok if there's no config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix("TS")

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	bindFlags(cmd, v)
	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		// // If using camelCase in the config file, replace hyphens with a camelCased string.
		// // Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		// if replaceHyphenWithCamelCase {
		// 	configName = strings.ReplaceAll(f.Name, "-", "")
		// }

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

type GlobalFlags struct {
	ClientId        string
	ClientSecret    string
	AuthKey         string
	Port            string
	Hostname        string
	Verbose         bool
	ConfigDirectory string
}

var gf = GlobalFlags{}

func Start() error {
	return rootCommand().Execute()
}

func rootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tailsys",
		Short: "A configuration manager built for Tailscale",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cmd)
		},
	}
	rootCmd.PersistentFlags().StringVar(&gf.ClientId, "client-id", "", "Oauth client id")
	viper.BindPFlag("client-id", rootCmd.PersistentFlags().Lookup("client-id"))
	rootCmd.PersistentFlags().StringVar(&gf.ClientSecret, "client-secret", "", "Oauth client client secret")
	viper.BindPFlag("client-secret", rootCmd.PersistentFlags().Lookup("client-secret"))
	rootCmd.MarkFlagsRequiredTogether("client-id", "client-secret")

	rootCmd.PersistentFlags().StringVar(&gf.AuthKey, "auth-key", "", "Pre-generated tailscale auth key")
	viper.BindPFlag("auth-key", rootCmd.PersistentFlags().Lookup("auth-key"))
	rootCmd.PersistentFlags().StringVarP(&gf.Port, "port", "p", "6655", "gRPC Port to listen on")
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	rootCmd.PersistentFlags().StringVar(&gf.Hostname, "hostname", "", "Override hostname")
	viper.BindPFlag("hostname", rootCmd.PersistentFlags().Lookup("hostname"))
	rootCmd.PersistentFlags().StringVar(&gf.ConfigDirectory, "data-directory", getConfigDirectory(), "Set the location for the data store")
	viper.BindPFlag("data-directory", rootCmd.PersistentFlags().Lookup("data-directory"))

	rootCmd.PersistentFlags().BoolVarP(&gf.Verbose, "verbose", "v", false, "Verbose logging")

	rootCmd.AddCommand(coodinationServerCommand())
	rootCmd.AddCommand(clientCommand())
	//	rootCmd.AddCommand(interactiveCommand())
	rootCmd.AddCommand(noninteractiveCommand())

	return rootCmd
}

type coFlags struct {
	DevMode bool
}

var cof = coFlags{}

func coodinationServerCommand() *cobra.Command {
	ccmd := &cobra.Command{
		Use:     "coordination-server",
		Aliases: []string{"co"},
		Short:   "Start the application coordination server",
		RunE: func(ccmd *cobra.Command, args []string) error {
			fmt.Println("starting the server code")

			var co coordination.Coordinator
			fmt.Println("dev-mode: ", cof.DevMode)
			fmt.Println("data-dir: ", gf.ConfigDirectory)

			ctx := context.Background()
			err := co.NewCoordinator(ctx,
				co.WithDevMode(cof.DevMode),
			)

			if err != nil {
				return err
			}

			hostname := gf.Hostname
			if err := co.Connect(ctx,
				co.WithAuthKey(gf.AuthKey),
				co.WithOauth(gf.ClientId, gf.ClientSecret),
				co.WithHostname(hostname),
				co.WithTags("tag:tailsys"),
				co.WithScopes("devices", "logs:read", "routes:read"),
				co.WithPort(gf.Port),
				co.WithConfigDir(gf.ConfigDirectory),
			); err != nil {
				return err
			}
			err = co.StartDatabase(ctx)
			if err != nil {
				return err
			}
			return co.StartRPCCoordinationServer(ctx)
		},
	}
	ccmd.Flags().BoolVar(&cof.DevMode, "dev", false, "Enable dev mode, accept all incoming keys")

	return ccmd
}

type clientFlags struct {
	DiscoveryTags      string
	CoordinationServer string
}

var cif = clientFlags{}

func clientCommand() *cobra.Command {
	ccmd := &cobra.Command{
		Use:     "client",
		Aliases: []string{"cl"},
		Short:   "Start the application in client mode",
		RunE: func(ccmd *cobra.Command, args []string) error {
			fmt.Println("starting the client code")
			ctx := context.Background()

			var cl client.Client
			err := cl.NewClient(ctx)
			if err != nil {
				return err
			}

			hostname := gf.Hostname
			coServer := cif.CoordinationServer
			if err := cl.Connect(ctx,
				cl.WithAuthKey(gf.AuthKey),
				cl.WithOauth(gf.ClientId, gf.ClientSecret),
				cl.WithHostname(hostname),
				cl.WithTags("tag:tailsys"),
				cl.WithScopes("devices", "logs:read", "routes:read"),
				cl.WithPort(gf.Port),
				cl.WithConfigDir(gf.ConfigDirectory),
			); err != nil {
				return err
			}
			fmt.Println("registered, starting client operations")
			err = cl.StartDatabase(ctx)
			if err != nil {
				return err
			}
			fmt.Println("connected, registering with coordination server")
			if err := cl.RegisterWithCoordinationServer(ctx, coServer); err != nil {
				return err
			}
			return cl.StartRPCClientMode(ctx)

		},
	}
	ccmd.Flags().StringVar(&cif.CoordinationServer, "coordination-server", "", "Hostname of the coordination server")
	ccmd.Flags().StringVar(&cif.DiscoveryTags, "discover-tags", "", "Tailnet tags to filter and discover hosts")
	return ccmd
}

// func interactiveCommand() *cobra.Command {
// 	ccmd := &cobra.Command{
// 		Use:     "interactive",
// 		Aliases: []string{"i"},
// 		Short:   "Start the application interactive ui",
// 		Run: func(ccmd *cobra.Command, args []string) {

// 		},
// 	}
// 	return ccmd
// }

type cmdFlags struct {
	Cmd                string
	Pattern            string
	CoordinationServer string
}

var cmdf = cmdFlags{}

func noninteractiveCommand() *cobra.Command {
	ccmd := &cobra.Command{
		Use:     "cmd",
		Aliases: []string{"cmd"},
		Short:   "Start the command non-interactively",
		// Run: func(ccmd *cobra.Command, args []string) {
		// 	fmt.Println("starting the non-interactive code")
  //     fmt.Println("coordination-server: ", cmdf.CoordinationServer)
		// },
	}
	ccmd.PersistentFlags().StringVar(&cmdf.CoordinationServer, "coordination-server", "", "Hostname of the coordination server")
	ccmd.PersistentFlags().StringVar(&cmdf.Pattern, "pattern", "p", "pattern of nodes to look up")
  ccmd.MarkFlagRequired("pattern")

  ccmd.AddCommand(getNodes())
  ccmd.AddCommand(sendCommandToNodes())
	return ccmd
}

func getNodes() *cobra.Command {
	ccmd := &cobra.Command{
		Use:     "get-nodes",
		Aliases: []string{"gn"},
		Short:   "Use a pattern to find nodes",
		RunE: func(ccmd *cobra.Command, args []string) error {
      fmt.Printf("getting nodes that match pattern: %s\n", cmdf.Pattern)
      var client commander.Client
      if err := client.NewClient(
        ccmd.Context(),
        client.WithCoordinationServer(cmdf.CoordinationServer),
        ); err != nil {
          return err
        }

			if err := client.ConnectCmd(ccmd.Context(),
				client.WithAuthKey(gf.AuthKey),
				client.WithOauth(gf.ClientId, gf.ClientSecret),
				client.WithHostname(gf.Hostname),
				client.WithTags("tag:tailsys"),
				client.WithScopes("devices", "logs:read", "routes:read"),
				client.WithPort(gf.Port),
				client.WithConfigDir(gf.ConfigDirectory),
			); err != nil {
				return err
			}

      if err := client.GetNodes(ccmd.Context(), cmdf.Pattern); err != nil {
        fmt.Println(fmt.Errorf("error getting nodes %w", err))
      }
      return nil
		},
	}
	return ccmd
}

func sendCommandToNodes() *cobra.Command {
	ccmd := &cobra.Command{
		Use:     "send-command",
		Aliases: []string{"sc"},
		Short:   "Use a pattern to send command to nodes",
		RunE: func(ccmd *cobra.Command, args []string) error {
      fmt.Printf("sending command: %s\n that match pattern %s\n", cmdf.Cmd, cmdf.Pattern)
      var client commander.Client
      if err := client.NewClient(
        ccmd.Context(),
        client.WithCoordinationServer(cmdf.CoordinationServer),
        ); err != nil {
          return err
        }

			if err := client.ConnectCmd(ccmd.Context(),
				client.WithAuthKey(gf.AuthKey),
				client.WithOauth(gf.ClientId, gf.ClientSecret),
				client.WithHostname(gf.Hostname),
				client.WithTags("tag:tailsys"),
				client.WithScopes("devices", "logs:read", "routes:read"),
				client.WithPort(gf.Port),
				client.WithConfigDir(gf.ConfigDirectory),
			); err != nil {
				return err
			}

      if err := client.SendCommand(ccmd.Context(), cmdf.Cmd, cmdf.Pattern); err != nil {
        fmt.Println(fmt.Errorf("error getting nodes %w", err))
      }

      return nil
		},
	}
  ccmd.Flags().StringVar(&cmdf.Cmd, "command", "c", "command to send to nodes")
  ccmd.MarkFlagRequired("command")

	return ccmd
}

func getConfigDirectory() string {
	ddir := ""
	if Check() {
		ddir = "/var/lib/tailsys/"
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			cobra.CheckErr(err)
		}
		ddir = filepath.Join(home, ".local", "tailsys")
	}
	return ddir
}
