package connections

import (
	"context"
	"fmt"
)

//TODO: This should all be simpler

// connect to the tailnet using oauth credentials
func (tn *Tailnet) ConnectOauth(ctx context.Context, id, secret, hostname string) error {
	fmt.Printf("connecting %s as with oauth\n", hostname)
	srv, err := tn.NewConnection(ctx,
		tn.WithOauth(id, secret),
		//TODO: This needs to be parameterized in the config
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
		tn.WithHostname(hostname),
	)

	if err != nil {
		return err
	}
	err = tn.createRPCServer(srv)
	if err != nil {
		return err
	}

	return nil
}

// connect to the tailnet using a pre-generated auth-key
func (tn *Tailnet) ConnectAuthKey(ctx context.Context, authKey, hostname string) error {
	srv, err := tn.NewConnection(ctx,
		tn.WithAuthKey(authKey),

		//TODO: This needs to be parameterized in the config
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
		tn.WithHostname(hostname),
	)

	if err != nil {
		return err
	}

	err = tn.createRPCServer(srv)
	if err != nil {
		return err
	}
	return nil
}
