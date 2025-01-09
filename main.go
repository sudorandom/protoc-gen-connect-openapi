package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"google.golang.org/protobuf/proto"
	pluginpb "google.golang.org/protobuf/types/pluginpb"

	"github.com/alecthomas/kong"
	"github.com/lmittmann/tint"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type CLI struct {
	Version VersionFlag `name:"version" help:"Print version information and quit"`
}

func main() {
	version := fullVersion()
	cli := CLI{
		Version: VersionFlag(version),
	}
	_ = kong.Parse(&cli,
		kong.Name("protoc-gen-connect-openapi"),
		kong.Description("Plugin for generating OpenAPIv3 from protobufs matching the Connect RPC interface."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{"version": version})

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
func fullVersion() string {
	return fmt.Sprintf("%s (%s) @ %s; %s", version, commit, date, runtime.Version())
}
