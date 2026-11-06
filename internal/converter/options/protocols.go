// Copyright 2023-2026 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

type Protocol struct {
	Name         string
	ContentType  string
	RequestDesc  string
	ResponseDesc string
	IsStreaming  bool
	IsBinary     bool
}

var Protocols = []Protocol{
	{
		// No need to explain JSON :)
		Name:        "json",
		ContentType: "application/json",
	},
	{
		Name:        "proto",
		ContentType: "application/proto",
		IsBinary:    true,
	},
	{
		Name:         "connect+json",
		ContentType:  "application/connect+json",
		RequestDesc:  "The request is JSON with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		ResponseDesc: "The response is JSON with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "connect+proto",
		ContentType:  "application/connect+proto",
		RequestDesc:  "The request is binary-encoded protobuf with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		ResponseDesc: "The response is binary-encoded protobuf with Connect protocol framing to support streaming RPCs. See the [Connect Protocol](https://connectrpc.com/docs/protocol) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc",
		ContentType:  "application/grpc",
		RequestDesc:  "The request is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc+proto",
		ContentType:  "application/grpc+proto",
		RequestDesc:  "The request is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc+json",
		ContentType:  "application/grpc+json",
		RequestDesc:  "The request is uses the gRPC protocol but with JSON encoding. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		ResponseDesc: "The response is uses the gRPC protocol but with JSON encoding. See the [the gRPC documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc-web",
		ContentType:  "application/grpc-web",
		RequestDesc:  "The request is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc-web+proto",
		ContentType:  "application/grpc-web+proto",
		RequestDesc:  "The request is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
	{
		Name:         "grpc-web+json",
		ContentType:  "application/grpc-web+json",
		RequestDesc:  "The request is uses the gRPC-Web protocol but with JSON encoding. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		ResponseDesc: "The response is uses the gRPC-Web protocol but with JSON encoding. See the [the gRPC-Web documentation](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md) for more.",
		IsStreaming:  true,
		IsBinary:     true,
	},
}
