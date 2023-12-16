package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pseudomuto/protokit/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// TestConvert calls greetings.Hello with a name, checking
// for a valid return value.
func TestConvert(t *testing.T) {
	paths, err := filepath.Glob("fixtures/*.proto")
	require.NoError(t, err)
	for _, protofile := range paths {
		protofile := protofile

		formats := []string{}
		for _, format := range []string{"yaml", "json"} {
			if _, err := os.Stat(makeOutputPath(protofile, format)); errors.Is(err, os.ErrNotExist) {
				continue
			}
			require.NoError(t, err)
			formats = append(formats, format)
		}

		require.NotZero(t, len(formats), "at least one output format has to exist")

		relPath := strings.TrimPrefix(protofile, "fixtures/")
		for _, format := range formats {
			format := format
			t.Run(protofile+"→"+format, func(t *testing.T) {

				// Make Generation Request
				pf, err := utils.LoadDescriptorSet("fixtures", "fileset.binpb")
				require.NoError(t, err)
				req := utils.CreateGenRequest(pf, relPath)
				params := fmt.Sprintf("format_%s", format)
				req.Parameter = &params

				b, err := proto.Marshal(req)
				require.NoError(t, err)

				// Call the conversion code!
				resp, err := ConvertFrom(bytes.NewBuffer(b))
				require.NoError(t, err)
				assert.Len(t, resp.File, 1)
				file := resp.File[0]
				assert.NotNil(t, file.Name)
				assert.Equal(t, strings.TrimSuffix(relPath, filepath.Ext(relPath))+".openapi."+format, file.GetName())

				// Load in our expected output and compare it against what we actually got
				expectedFile, err := os.ReadFile(makeOutputPath(protofile, format))
				require.NoError(t, err)
				assert.Equal(t, string(expectedFile), file.GetContent())

				// Validate
				var validate validator.Validator
				t.Run("validate", func(tt *testing.T) {
					document, err := libopenapi.NewDocument([]byte(file.GetContent()))
					require.NoError(t, err)
					var errs []error
					validate, errs = validator.NewValidator(document)
					require.Len(t, errs, 0, errs)
				})

				// Load in validation test cases to check OpenAPI specifications against sample requests
				testCaseFilePath := strings.TrimSuffix(protofile, filepath.Ext(protofile)) + ".cases.yaml"
				if _, err := os.Stat(testCaseFilePath); errors.Is(err, os.ErrNotExist) {
					t.Logf("No cases file: %+s", testCaseFilePath)
					return
				}

				testCaseFileBytes, err := os.ReadFile(testCaseFilePath)
				require.NoError(t, err)

				testCaseFile := &TestCaseFile{}
				if err := yaml.Unmarshal(testCaseFileBytes, testCaseFile); err != nil {
					require.NoError(t, err)
				}

				for _, testCase := range testCaseFile.Cases {
					testCase := testCase
					if testCase.Method == "" {
						testCase.Method = "POST"
					}
					if len(testCase.Errors) == 0 {
						testCase.Errors = []string{}
					}
					t.Run(testCase.Name, func(tt *testing.T) {
						var body io.Reader
						if len(testCase.Body) > 0 {
							body = strings.NewReader(testCase.Body)
						}
						req, err := http.NewRequest(testCase.Method, testCase.Path, body)
						for k, v := range testCase.Headers {
							req.Header.Add(k, v)
						}
						require.NoError(tt, err)

						ok, errs := validate.ValidateHttpRequest(req)
						require.Len(tt, errs, len(testCase.Errors), "Incorrect number of errors", errs)

						for i, err := range errs {
							assert.Regexp(tt, testCase.Errors[i], err.Error())
						}
						assert.Equal(tt, ok, len(testCase.Errors) == 0)
					})
				}
			})
		}
	}
}

type TestCaseFile struct {
	Cases []TestCase `yaml:"cases"`
}

type TestCase struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
	Errors  []string          `yaml:"errors"`
}

func makeOutputPath(protofile, format string) string {
	return strings.TrimSuffix(protofile, filepath.Ext(protofile)) + ".openapi." + format
}
