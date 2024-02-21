# tailsys
Remote command runner written in golang


## Compiling Protocol Buffers
```bash
protoc -I=./protos --go_out=./ protos/*.proto --go-grpc_out=./
```
