// Package codec registers a JSON codec as the default gRPC codec.
// This lets us use plain Go structs for gRPC messages instead of requiring
// protoc-generated code. In production, swap this for protobuf for performance.
package codec

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

func init() {
	// Registering with name "proto" replaces the default protobuf codec.
	encoding.RegisterCodec(JSONCodec{})
}

// JSONCodec implements the gRPC encoding.Codec interface using encoding/json.
type JSONCodec struct{}

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Name returns "proto" so it replaces the default gRPC codec.
func (JSONCodec) Name() string { return "proto" }
