// Package codec registers a JSON codec as the default gRPC codec.
package codec

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// JSONCodec implements the gRPC encoding.Codec interface using encoding/json.
type JSONCodec struct{}

func (JSONCodec) Marshal(v interface{}) ([]byte, error)      { return json.Marshal(v) }
func (JSONCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
func (JSONCodec) Name() string                               { return "proto" }

// Register installs the JSON codec as the active gRPC codec.
// Calls this as the FIRST line in every service's main() — before any grpc.NewServer or grpc.Dial.
func Register() {
	encoding.RegisterCodec(JSONCodec{})
}
