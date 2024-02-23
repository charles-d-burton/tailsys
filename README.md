# Tailsys 
Remote command runner written in golang

## Requirements
This software relies on using [Tailscale](https://tailscale.com), it cannot work without this software.
You must also setup either an [auth-key](https://tailscale.com/kb/1085/auth-keys) or preferably configure your tailnet for [oauth](https://tailscale.com/kb/1215/oauth-clients)

## Testing with Docker Compose

## Compiling Protocol Buffers
```bash
protoc -I=./protos --go_out=./ protos/*.proto --go-grpc_out=./
```
