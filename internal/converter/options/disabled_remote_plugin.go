//go:build remote_plugin

package options

import "fmt"

const optionDisabledRemotePlugin = `The '%s' option is disabled when ran with a remote plugin. If you need this option, please use the local plugin instead or consider using gnostic OpenAPIv3 annotations in your protobuf.

See here for more information: https://github.com/sudorandom/protoc-gen-connect-openapi/blob/main/remote-plugin.md`

var disabledOptions = map[string]string{
	"base":     fmt.Sprintf(optionDisabledRemotePlugin, "base"),
	"override": fmt.Sprintf(optionDisabledRemotePlugin, "override"),
}
