package converter_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pseudomuto/protokit/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

var scenarios = []Scenario{
	{Name: "standard", Options: "allow-get,with-streaming"},
	{Name: "proto_names", Options: "with-proto-names"},
	{Name: "trim_unused_type", Options: "trim-unused-types"},
	{Name: "with_base", Options: "base=testdata/with_base/base.yaml,trim-unused-types"},
}

type Scenario struct {
	Name    string
	Options string
}

func generateAndCheckResult(t *testing.T, options, format, protofile string) string {
	relPath := path.Join("internal", "converter", protofile)

	// Make Generation Request
	pf, err := utils.LoadDescriptorSet("testdata", "fileset.binpb")
	require.NoError(t, err)
	req := utils.CreateGenRequest(pf, relPath)
	var sb strings.Builder
	sb.WriteString("debug,format=")
	sb.WriteString(format)
	if len(options) > 0 {
		sb.WriteString(",")
		sb.WriteString(options)
	}
	req.Parameter = proto.String(sb.String())

	b, err := proto.Marshal(req)
	require.NoError(t, err)

	// Call the conversion code!
	resp, err := converter.ConvertFrom(bytes.NewBuffer(b))
	require.NoError(t, err)
	require.Len(t, resp.File, 1)
	file := resp.File[0]
	assert.NotNil(t, file.Name)
	assert.Equal(t, strings.TrimSuffix(relPath, filepath.Ext(relPath))+".openapi."+format, file.GetName())

	// Load in our expected output and compare it against what we actually got
	outputPath := makeOutputPath(protofile, format)
	_, statErr := os.Stat(outputPath)
	switch {
	case errors.Is(statErr, os.ErrNotExist):
		assert.NoError(t, os.WriteFile(outputPath, []byte(file.GetContent()), 0644))
	case statErr != nil:
		require.NoError(t, statErr)
	default:
		expectedFile, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Equal(t, string(expectedFile), file.GetContent())
	}
	return file.GetContent()
}

func validateOpenAPISpec(t *testing.T, protofile string, spec string) {
	config := datamodel.DocumentConfiguration{
		IgnorePolymorphicCircularReferences: true,
		IgnoreArrayCircularReferences:       true,
		SkipCircularReferenceCheck:          true,
	}

	document, err := libopenapi.NewDocumentWithConfiguration([]byte(spec), &config)
	require.NoError(t, err)

	var errs []error
	validate, errs := validator.NewValidator(document)
	require.Len(t, errs, 0, errs)

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
			path := testCase.Path
			if len(testCase.Query) > 0 {
				path += "?" + testCase.Query
			}
			req, err := http.NewRequest(testCase.Method, path, body)
			for k, v := range testCase.Headers {
				req.Header.Add(k, v)
			}
			require.NoError(tt, err)

			ok, errs := validate.ValidateHttpRequest(req)
			require.Len(tt, errs, len(testCase.Errors), "Incorrect number of errors: %+v", errs)

			for i, err := range errs {
				assert.Regexp(tt, testCase.Errors[i], err.Error())
			}
			assert.Equal(tt, ok, len(testCase.Errors) == 0)
		})
	}
}

// TestConvert uses data in testdata/ to make requests to generate openapi documents,
// checks if they are valid OpenAPI documents and validates that a list of example
// requests would conform to the OpenAPI spec or fail.
func TestConvert(t *testing.T) {
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			paths, err := filepath.Glob("testdata/" + scenario.Name + "/**.proto")
			require.NoError(t, err)
			for _, protofile := range paths {
				protofile := protofile

				formats := []string{"yaml", "json"}

				for _, format := range formats {
					format := format
					t.Run(path.Base(protofile)+"→"+format, func(t *testing.T) {
						spec := generateAndCheckResult(t, scenario.Options, format, protofile)
						// Validate
						t.Run("validate", func(tt *testing.T) {
							validateOpenAPISpec(t, protofile, spec)
						})
					})
				}
			}
		})
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
	Query   string            `yaml:"query"`
	Errors  []string          `yaml:"errors"`
}

func makeOutputPath(protofile, format string) string {
	dir, file := filepath.Split(strings.TrimSuffix(protofile, filepath.Ext(protofile)) + ".openapi." + format)
	return filepath.Join(dir, "output", file)
}

func TestTrimUnusedTypes(t *testing.T) {
	testCases := []struct {
		name           string
		input         string
		expectedTypes []string
		excludedTypes []string
	}{
		{
			name: "simple_unused_type",
			input: `
openapi: 3.1.0
components:
  schemas:
    Used:
      type: object
      properties:
        name:
          type: string
    Unused:
      type: object
      properties:
        description:
          type: string
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Used'
`,
			expectedTypes: []string{"Used"},
			excludedTypes: []string{"Unused"},
		},
		{
			name: "nested_references",
			input: `
openapi: 3.1.0
components:
  schemas:
    Parent:
      type: object
      properties:
        child:
          $ref: '#/components/schemas/Child'
    Child:
      type: object
      properties:
        name:
          type: string
    Unused:
      type: object
      properties:
        description:
          type: string
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Parent'
`,
			expectedTypes: []string{"Parent", "Child"},
			excludedTypes: []string{"Unused"},
		},
		{
			name: "circular_reference",
			input: `
openapi: 3.1.0
components:
  schemas:
    Parent:
      type: object
      properties:
        child:
          $ref: '#/components/schemas/Child'
    Child:
      type: object
      properties:
        parent:
          $ref: '#/components/schemas/Parent'
    Unused:
      type: object
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Parent'
`,
			expectedTypes: []string{"Parent", "Child"},
			excludedTypes: []string{"Unused"},
		},
		{
			name: "array_references",
			input: `
openapi: 3.1.0
components:
  schemas:
    Collection:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/Item'
    Item:
      type: object
      properties:
        name:
          type: string
    Unused:
      type: object
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Collection'
`,
			expectedTypes: []string{"Collection", "Item"},
			excludedTypes: []string{"Unused"},
		},
		{
			name: "composition_references",
			input: `
openapi: 3.1.0
components:
  schemas:
    Combined:
      allOf:
        - $ref: '#/components/schemas/Part1'
        - $ref: '#/components/schemas/Part2'
    Part1:
      type: object
      properties:
        name:
          type: string
    Part2:
      type: object
      properties:
        age:
          type: integer
    Unused:
      type: object
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Combined'
`,
			expectedTypes: []string{"Combined", "Part1", "Part2"},
			excludedTypes: []string{"Unused"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the input OpenAPI spec
			doc, err := libopenapi.NewDocument([]byte(tc.input))
			require.NoError(t, err)
			
			model, errs := doc.BuildV3Model()
			require.Empty(t, errs)

			// Apply the trimming
			err = converter.TrimUnusedTypes(&model.Model)
			require.NoError(t, err)

			// Check that expected types are present
			for _, expectedType := range tc.expectedTypes {
				_, ok := model.Model.Components.Schemas.Get(expectedType)
				assert.True(t, ok, "Expected type %s should be present", expectedType)
			}

			// Check that excluded types are not present
			for _, excludedType := range tc.excludedTypes {
				_, ok := model.Model.Components.Schemas.Get(excludedType)
				assert.False(t, ok, "Excluded type %s should not be present", excludedType)
			}
		})
	}
}