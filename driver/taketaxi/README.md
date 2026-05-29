# taketaxi

Microservice project.

## Options

- **Protocol**: grpc
- **HTTP**: gin
- **IDL**: proto

## Structure
```
taketaxi/
├── bffDriver/
├── srvDriver/
├── common/
├── pkg/
└── scripts/
```

## Build
```bash
go mod init github.com/yourorg/taketaxi
./scripts/gen_proto.sh
./scripts/build.sh
```
