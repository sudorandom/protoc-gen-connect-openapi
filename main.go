package main

import (
	"fmt"
	"log/slog"
	"os"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/proto"
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{}),
	))

	resp, err := ConvertFrom(os.Stdin)
	if err != nil {
		message := fmt.Sprintf("Failed to read input: %v", err)
		slog.Error(message)
		renderResponse(&plugin.CodeGeneratorResponse{
			Error: &message,
		})
		os.Exit(1)
	}

	renderResponse(resp)
}

func renderResponse(resp *plugin.CodeGeneratorResponse) {
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
