package kcl_parse_file_issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kcl-lang.io/kcl-go/pkg/native"
	"kcl-lang.io/kcl-go/pkg/service"
	"kcl-lang.io/kcl-go/pkg/spec/gpyrpc"
)

func Test(t *testing.T) {
	testCases := []struct {
		Name   string
		Client service.KclvmService
	}{
		{
			Name:   "pygrpc",
			Client: service.NewKclvmServiceClient(),
		},
		{
			Name:   "native",
			Client: native.NewNativeServiceClient(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Run("simple", func(t *testing.T) {
				source := `
import units

data1 = 32 * units.Ki
`
				result, err := testCase.Client.ParseFile(&gpyrpc.ParseFile_Args{
					Path:   "source.k",
					Source: source,
				})
				if err != nil {
					require.Fail(t, "ParseFile returns error: %s", trimString(err.Error()))
				}
				assert.NotEmpty(t, result.AstJson)
				require.Len(t, result.Errors, 0)
			})
			t.Run("pkgpath not found", func(t *testing.T) {
				source := `
import units
import data.cloud as cloud_pkg

data1 = 32 * units.Ki
Data2 = 42 * units.Ki * cloud_pkg.Foo

lambda1 = lambda x: int, y: int -> int {
    x - y
}
Lambda2 = lambda x: int, y: int -> int {
    x + y
}`
				result, err := testCase.Client.ParseFile(&gpyrpc.ParseFile_Args{
					Path:   "source.k",
					Source: source,
				})
				if err != nil {
					require.Fail(t, "ParseFile returns error: %s", trimString(err.Error()))
				}
				assert.NotEmpty(t, result.AstJson)
				var errorMsgs []string
				for _, protoErr := range result.Errors {
					for _, msg := range protoErr.Messages {
						errorMsgs = append(errorMsgs, msg.Msg)
					}
				}
				require.Equal(t, errorMsgs, []string{
					`pkgpath data.cloud not found in the program`,
					`try 'kcl mod add data' to download the package not found`,
					`find more package on 'https://artifacthub.io'`,
				})
			})
		})
	}
}

func trimString(v string) string {
	if len(v) > 100 {
		return v[:100]
	}
	return v
}
