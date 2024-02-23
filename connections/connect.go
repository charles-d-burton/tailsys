package connections

import (
	"context"

)

//connect to the tailnet using oauth credentials
func ConnectOauth(ctx context.Context, id, secret string) (*Tailnet, error) {
	var tn Tailnet
	srv, err := tn.NewConnection(ctx,
		tn.WithOauth(id, secret),
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
	)

	if err != nil {
		return nil, err
	}
  err = tn.createRPCServer(ctx, srv)
  if err != nil {
    return nil, err
  }

	return &tn, nil
}

//connect to the tailnet using a pre-generated auth-key
func ConnectAuthKey(ctx context.Context, authKey string) (*Tailnet, error) {
	var tn Tailnet
	srv, err := tn.NewConnection(ctx,
		tn.WithAuthKey(authKey),

    //TODO: This needs to be parameterized in the config
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
	)

	if err != nil {
		return nil, err
	}

  err = tn.createRPCServer(ctx, srv)
  if err != nil {
    return nil, err
  }
	return &tn, nil
}


