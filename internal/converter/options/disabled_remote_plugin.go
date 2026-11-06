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

//go:build remote_plugin

package options

import "fmt"

const optionDisabledRemotePlugin = `The '%s' option is disabled when ran with a remote plugin. If you need this option, please use the local plugin instead or consider using gnostic OpenAPIv3 annotations in your protobuf.

See here for more information: https://github.com/connectrpc/protoc-gen-connect-openapi/blob/main/remote-plugin.md`

var disabledOptions = map[string]string{
	"base":     fmt.Sprintf(optionDisabledRemotePlugin, "base"),
	"override": fmt.Sprintf(optionDisabledRemotePlugin, "override"),
}
