module backend-diplomski/tools/interaction-simulator

go 1.25.8

require (
	backend-diplomski/proto v0.0.0
	backend-diplomski/shared v0.0.0
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.83.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace backend-diplomski/proto => ../../proto

replace backend-diplomski/shared => ../../shared
