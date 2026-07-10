package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/util"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/visibility"
	yaml "go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type AsyncAPISpec struct {
	AsyncAPI   string                        `json:"asyncapi" yaml:"asyncapi"`
	Info       AsyncAPIInfo                  `json:"info" yaml:"info"`
	Channels   map[string]*AsyncAPIChannel   `json:"channels" yaml:"channels"`
	Operations map[string]*AsyncAPIOperation `json:"operations" yaml:"operations"`
	Components *AsyncAPIComponents           `json:"components,omitempty" yaml:"components,omitempty"`
}

type AsyncAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type AsyncAPIChannel struct {
	Address     string                      `json:"address" yaml:"address"`
	Messages    map[string]*AsyncAPIMessage `json:"messages,omitempty" yaml:"messages,omitempty"`
	Description string                      `json:"description,omitempty" yaml:"description,omitempty"`
}

type AsyncAPIMessage struct {
	Name        string         `json:"name" yaml:"name"`
	Payload     map[string]any `json:"payload" yaml:"payload"`
	Title       string         `json:"title,omitempty" yaml:"title,omitempty"`
	Summary     string         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
}

type AsyncAPIOperation struct {
	Action      string         `json:"action" yaml:"action"` // "send" or "receive"
	Channel     *AsyncAPIRef   `json:"channel" yaml:"channel"`
	Messages    []*AsyncAPIRef `json:"messages,omitempty" yaml:"messages,omitempty"`
	Summary     string         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
}

type AsyncAPIRef struct {
	Ref string `json:"$ref" yaml:"$ref"`
}

type AsyncAPIComponents struct {
	Schemas map[string]any `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}

func GenerateAsyncAPI(opts options.Options, fds []protoreflect.FileDescriptor) (string, error) {
	spec := &AsyncAPISpec{
		AsyncAPI: "3.1.0",
		Info: AsyncAPIInfo{
			Title:   "Streaming API",
			Version: "1.0.0",
		},
		Channels:   make(map[string]*AsyncAPIChannel),
		Operations: make(map[string]*AsyncAPIOperation),
		Components: &AsyncAPIComponents{
			Schemas: make(map[string]any),
		},
	}

	dummyDoc := &v3.Document{
		Version: "3.1.0",
		Info: &base.Info{
			Title:   "Dummy",
			Version: "1.0.0",
		},
		Paths: &v3.Paths{
			PathItems: orderedmap.New[string, *v3.PathItem](),
		},
		Components: &v3.Components{
			Schemas: orderedmap.New[string, *base.SchemaProxy](),
		},
	}

	hasStreaming := false
	var titleParts []string
	var descParts []string

	for _, fd := range fds {
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			if !opts.HasService(service.FullName()) {
				continue
			}
			if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(service), opts.AllowedVisibilities) {
				continue
			}

			titleParts = append(titleParts, string(service.Name()))
			serviceLoc := fd.SourceLocations().ByDescriptor(service)
			if serviceComments := util.FormatComments(serviceLoc); serviceComments != "" {
				descParts = append(descParts, serviceComments)
			}

			methods := service.Methods()
			for j := 0; j < methods.Len(); j++ {
				method := methods.Get(j)
				if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(method), opts.AllowedVisibilities) {
					continue
				}

				isStream := method.IsStreamingClient() || method.IsStreamingServer()
				if !isStream {
					continue
				}
				hasStreaming = true

				methodLoc := fd.SourceLocations().ByDescriptor(method)
				summary, description := util.FormatOperationComments(methodLoc)
				if summary == "" {
					summary = string(method.Name())
				}

				pkg := string(service.FullName().Parent())
				channelAddress := formatChannelPath(opts.AsyncAPIChannelTemplate, pkg, string(service.Name()), string(method.Name()))

				channelKey := string(service.Name()) + string(method.Name()) + "Channel"
				requestMessageKey := string(method.Input().Name())
				responseMessageKey := string(method.Output().Name())

				AddMessageSchemas(opts, method.Input(), dummyDoc)
				AddMessageSchemas(opts, method.Output(), dummyDoc)

				channel := &AsyncAPIChannel{
					Address:     channelAddress,
					Description: description,
					Messages:    make(map[string]*AsyncAPIMessage),
				}

				channel.Messages[requestMessageKey] = &AsyncAPIMessage{
					Name:        requestMessageKey,
					Title:       string(method.Input().Name()),
					Description: util.FormatComments(fd.SourceLocations().ByDescriptor(method.Input())),
					Payload: map[string]any{
						"$ref": fmt.Sprintf("#/components/schemas/%s", string(method.Input().FullName())),
					},
				}

				channel.Messages[responseMessageKey] = &AsyncAPIMessage{
					Name:        responseMessageKey,
					Title:       string(method.Output().Name()),
					Description: util.FormatComments(fd.SourceLocations().ByDescriptor(method.Output())),
					Payload: map[string]any{
						"$ref": fmt.Sprintf("#/components/schemas/%s", string(method.Output().FullName())),
					},
				}

				spec.Channels[channelKey] = channel

				sendOpKey := string(service.Name()) + string(method.Name()) + "Send"
				spec.Operations[sendOpKey] = &AsyncAPIOperation{
					Action:      "send",
					Summary:     fmt.Sprintf("Send: %s", summary),
					Description: description,
					Channel:     &AsyncAPIRef{Ref: fmt.Sprintf("#/channels/%s", channelKey)},
					Messages: []*AsyncAPIRef{
						{Ref: fmt.Sprintf("#/channels/%s/messages/%s", channelKey, requestMessageKey)},
					},
				}

				receiveOpKey := string(service.Name()) + string(method.Name()) + "Receive"
				spec.Operations[receiveOpKey] = &AsyncAPIOperation{
					Action:      "receive",
					Summary:     fmt.Sprintf("Receive: %s", summary),
					Description: description,
					Channel:     &AsyncAPIRef{Ref: fmt.Sprintf("#/channels/%s", channelKey)},
					Messages: []*AsyncAPIRef{
						{Ref: fmt.Sprintf("#/channels/%s/messages/%s", channelKey, responseMessageKey)},
					},
				}
			}
		}
	}

	if len(titleParts) > 0 {
		spec.Info.Title = strings.Join(titleParts, ", ") + " WebSockets"
	}
	if len(descParts) > 0 {
		spec.Info.Description = strings.Join(descParts, "\n\n")
	}

	if !hasStreaming {
		return "", nil
	}

	jsonBytes, err := dummyDoc.RenderJSON("  ")
	if err != nil {
		return "", fmt.Errorf("failed to render schemas: %w", err)
	}

	var parsedDoc struct {
		Components struct {
			Schemas map[string]any `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(jsonBytes, &parsedDoc); err != nil {
		return "", fmt.Errorf("failed to parse rendered schemas: %w", err)
	}

	spec.Components.Schemas = parsedDoc.Components.Schemas

	var outContent []byte
	if opts.Format == "json" {
		outContent, err = json.MarshalIndent(spec, "", "  ")
	} else {
		outContent, err = yaml.Marshal(spec)
	}
	if err != nil {
		return "", fmt.Errorf("failed to marshal AsyncAPI spec: %w", err)
	}

	return string(outContent), nil
}

func formatChannelPath(template string, pkg string, service string, method string) string {
	res := template
	if pkg == "" {
		res = strings.ReplaceAll(res, "{package}.", "")
		res = strings.ReplaceAll(res, "{package}", "")
	} else {
		res = strings.ReplaceAll(res, "{package}", pkg)
	}
	res = strings.ReplaceAll(res, "{service}", service)
	res = strings.ReplaceAll(res, "{method}", method)
	return res
}
