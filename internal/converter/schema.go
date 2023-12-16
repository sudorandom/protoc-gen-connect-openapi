package converter

import (
	"log/slog"

	"github.com/swaggest/jsonschema-go"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type State struct {
	CurrentFile      protoreflect.FileDescriptor
	ExternalMessages []protoreflect.MessageDescriptor
	ExternalEnums    []protoreflect.EnumDescriptor
}

func enumToSchema(state *State, tt protoreflect.EnumDescriptor) *jsonschema.Schema {
	slog.Info("enumToSchema", slog.Any("descriptor", tt.FullName()))
	s := &jsonschema.Schema{}
	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(formatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	s.WithType(jsonschema.String.Type())
	children := []interface{}{}
	values := tt.Values()
	for i := 0; i < values.Len(); i++ {
		value := values.Get(i)
		children = append(children, string(value.Name()))
		children = append(children, value.Number())
	}
	s.WithEnum(children)
	return s
}

func messageToSchema(state *State, tt protoreflect.MessageDescriptor) *jsonschema.Schema {
	slog.Info("messageToSchema", slog.Any("descriptor", tt.FullName()))
	if isWellKnown(tt) {
		return wellKnownToSchema(tt)
	}
	s := &jsonschema.Schema{}
	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(formatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	s.WithType(jsonschema.Object.Type())

	fields := tt.Fields()
	children := make(map[string]jsonschema.SchemaOrBool, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		child := fieldToSchema(state, field)
		children[field.JSONName()] = jsonschema.SchemaOrBool{TypeObject: child}
	}

	s.WithProperties(children)
	return s
}

func fieldToSchema(state *State, tt protoreflect.FieldDescriptor) *jsonschema.Schema {
	slog.Info("fieldToSchema", slog.Any("descriptor", tt.FullName()))
	s := &jsonschema.Schema{}

	switch tt.Kind() {
	case protoreflect.BoolKind:
		s.WithType(jsonschema.Boolean.Type())
	case protoreflect.EnumKind:
		if tt.Enum().ParentFile().Path() != state.CurrentFile.Path() {
			state.ExternalEnums = append(state.ExternalEnums, tt.Enum())
		}
		s.WithRef("#/components/schemas/" + string(tt.Enum().FullName()))
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind:
		s.WithType(jsonschema.Integer.Type())
	case protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		s.WithType(jsonschema.Integer.Type())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.FloatKind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.DoubleKind:
		s.WithType(jsonschema.Number.Type())
	case protoreflect.StringKind:
		s.WithType(jsonschema.String.Type())
	case protoreflect.BytesKind:
		s.WithType(jsonschema.String.Type())
	case protoreflect.MessageKind:
		if tt.Message().ParentFile().Path() != state.CurrentFile.Path() {
			state.ExternalMessages = append(state.ExternalMessages, tt.Message())
		}
		s.WithRef("#/components/schemas/" + string(tt.Message().FullName()))
		s.WithType(jsonschema.Object.Type())
	}

	// Handle maps
	if tt.IsMap() {
		s.AdditionalProperties = &jsonschema.SchemaOrBool{TypeObject: fieldToSchema(state, tt.MapValue())}
		s.WithType(jsonschema.Object.Type())
		s.Ref = nil
	}

	// Handle Lists
	if tt.IsList() {
		wrapped := s
		s = &jsonschema.Schema{}
		s.WithType(jsonschema.Array.Type())
		s.WithItems(jsonschema.Items{SchemaArray: []jsonschema.SchemaOrBool{{TypeObject: wrapped}}})
	}

	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(formatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	return s
}

func fileToSchema(state *State, tt protoreflect.FileDescriptor) *jsonschema.Schema {
	slog.Info("fileOfToSchema", slog.Any("descriptor", tt.FullName()))
	state.CurrentFile = tt
	s := &jsonschema.Schema{}
	s.WithID(string(tt.FullName()))
	s.WithTitle(string(tt.Name()))
	s.WithDescription(formatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	children := []jsonschema.SchemaOrBool{}
	enums := tt.Enums()
	for i := 0; i < enums.Len(); i++ {
		child := enumToSchema(state, enums.Get(i))
		children = append(children, jsonschema.SchemaOrBool{TypeObject: child})
	}
	messages := tt.Messages()
	for i := 0; i < messages.Len(); i++ {
		message := messages.Get(i)
		child := messageToSchema(state, message)
		children = append(children, jsonschema.SchemaOrBool{TypeObject: child})

		// messages can also have enums defined in them. This is handled by adding them to the root
		// schema with a fully-qualified path
		enums := message.Enums()
		for i := 0; i < enums.Len(); i++ {
			child := enumToSchema(state, enums.Get(i))
			children = append(children, jsonschema.SchemaOrBool{TypeObject: child})
		}
	}

	for _, child := range resolveExternalDescriptors(state) {
		children = append(children, jsonschema.SchemaOrBool{TypeObject: child})
	}

	s.WithItems(jsonschema.Items{SchemaArray: children})
	return s
}

func resolveExternalDescriptors(state *State) []*jsonschema.Schema {
	if len(state.ExternalEnums) == 0 && len(state.ExternalMessages) == 0 {
		return []*jsonschema.Schema{}
	}

	children := []*jsonschema.Schema{}
	for _, enum := range state.ExternalEnums {
		childState := &State{
			CurrentFile: enum.ParentFile(),
		}
		children = append(children, enumToSchema(childState, enum))
		children = append(children, resolveExternalDescriptors(childState)...)
	}

	for _, message := range state.ExternalMessages {
		childState := &State{
			CurrentFile: message.ParentFile(),
		}
		children = append(children, messageToSchema(childState, message))
		children = append(children, resolveExternalDescriptors(childState)...)
	}

	return children
}
