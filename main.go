package main

import (
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/protobuf/proto"
	pluginpb "google.golang.org/protobuf/types/pluginpb"

	"github.com/lmittmann/tint"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter"
)

func main() {
	resp, err := converter.ConvertFrom(os.Stdin)
	if err != nil {
		message := fmt.Sprintf("Failed to read input: %v", err)
		slog.Error(message)
		renderResponse(&pluginpb.CodeGeneratorResponse{
			Error: &message,
		})
		os.Exit(1)
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	renderResponse(resp)
}

func renderResponse(resp *pluginpb.CodeGeneratorResponse) {
	data, err := proto.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal response", slog.Any("error", err))
		return
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		slog.Error("failed to write response", slog.Any("error", err))
		return
	}
}
