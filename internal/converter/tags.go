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

package converter

import (
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/options"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/util"
	"connectrpc.com/protoc-gen-connect-openapi/internal/converter/visibility"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func fileToTags(opts options.Options, fd protoreflect.FileDescriptor) []*base.Tag {
	if opts.WithoutDefaultTags {
		return nil
	}
	tags := []*base.Tag{}
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if !opts.HasService(service.FullName()) {
			continue
		}
		if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(service), opts.AllowedVisibilities) {
			continue
		}
		loc := fd.SourceLocations().ByDescriptor(service)
		description := util.FormatComments(loc)

		tagName := string(service.FullName())
		if opts.ShortServiceTags {
			tagName = string(service.Name())
		}
		tags = append(tags, &base.Tag{
			Name:        tagName,
			Description: description,
		})
	}
	return tags
}
