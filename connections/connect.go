package connections

import (
	"github.com/urfave/cli/v2"
)

//connect to the tailnet using oauth credentials
func ConnectOauth(ctx *cli.Context, id, secret, hostname string) (*Tailnet, error) {
	var tn Tailnet
	srv, err := tn.NewConnection(ctx.Context,
		tn.WithOauth(id, secret),
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
    tn.WithHostname(hostname),
	)

	if err != nil {
		return nil, err
	}
  err = tn.createRPCServer(ctx.Context, srv)
  if err != nil {
    return nil, err
  }

	return &tn, nil
}

//connect to the tailnet using a pre-generated auth-key
func ConnectAuthKey(ctx *cli.Context, authKey, hostname string) (*Tailnet, error) {
	var tn Tailnet
	srv, err := tn.NewConnection(ctx.Context,
		tn.WithAuthKey(authKey),

    //TODO: This needs to be parameterized in the config
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
    tn.WithHostname(hostname),
	)

	if err != nil {
		return nil, err
	}

  err = tn.createRPCServer(ctx.Context, srv)
  if err != nil {
    return nil, err
  }
	return &tn, nil
}

