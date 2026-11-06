package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"

	"google.golang.org/protobuf/proto"
	pluginpb "google.golang.org/protobuf/types/pluginpb"

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

	if err := run(os.Stdin, os.Stdout); err != nil {
		slog.Error("failed to write plugin response", slog.Any("error", err))
		os.Exit(1)
	}
}

// run converts a CodeGeneratorRequest into a response. Conversion errors are
// reported inside the response, which is how the plugin protocol carries them
// back to protoc or buf with the actual message; only failing to write the
// response is returned.
func run(stdin io.Reader, stdout io.Writer) error {
	resp, err := converter.ConvertFrom(stdin)
	if err != nil {
		message := err.Error()
		resp = &pluginpb.CodeGeneratorResponse{Error: &message}
	}
	return renderResponse(stdout, resp)
}

func renderResponse(w io.Writer, resp *pluginpb.CodeGeneratorResponse) error {
	data, err := proto.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}

func fullVersion() string {
	version, commit, date := getVersionInfo()
	return fmt.Sprintf("%s (%s) @ %s; %s", version, commit, date, runtime.Version())
}
