## Generate fo v1
```bash

# install protoc
go mod tidy
go install \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
  google.golang.org/protobuf/cmd/protoc-gen-go \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc

protoc -I . \
  --go_out ./ --go_opt paths=source_relative \
  --go-grpc_out ./ --go-grpc_opt paths=source_relative \
  --grpc-gateway_out ./ \
  --grpc-gateway_opt logtostderr=true \
  --grpc-gateway_opt paths=source_relative \
  --grpc-gateway_opt generate_unbound_methods=true \
  --grpc-gateway_opt allow_repeated_fields_in_body=true \
  --openapiv2_out ./ \
  --openapiv2_opt allow_repeated_fields_in_body=true \
  --openapiv2_opt logtostderr=true \
  v1/market.proto




```