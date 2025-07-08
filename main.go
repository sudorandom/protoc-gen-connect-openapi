package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"

	"google.golang.org/protobuf/proto"
	pluginpb "google.golang.org/protobuf/types/pluginpb"

	"github.com/lmittmann/tint"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter"
)

func getVersionInfo() (version, commit, date string) {
	version = "dev"
	commit = "none"
	date = "unknown"

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			version = info.Main.Version
		}

		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) >= 7 {
					commit = setting.Value[:7] // Short commit hash
				} else {
					commit = setting.Value
				}
			case "vcs.time":
				date = setting.Value
			}
		}
	}

	return version, commit, date
}

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-connect-openapi %s\n", fullVersion())
		return
	}

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
	version, commit, date := getVersionInfo()
	return fmt.Sprintf("%s (%s) @ %s; %s", version, commit, date, runtime.Version())
}
