package cnbshim_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExports(t *testing.T) {
	tcs := []struct {
		name     string
		environ  []string // optional environ for the test env
		wantEnvs map[string]string
	}{
		{
			name: "export_simple",
			wantEnvs: map[string]string{
				"FOO": "bar",
			},
		},
		{
			name: "export_quotes",
			wantEnvs: map[string]string{
				"FOO": `b"a"r`,
			},
		},
		{
			name: "export_script",
			environ: []string{
				"PATH=/usr/bin:/bin:/custom",
			},
			wantEnvs: map[string]string{
				"PATH":     "/some-path:/another:/usr/bin:/bin:/custom:/end",
				"SOME_VAR": "yes",
				"SQUARE_1": "1",
				"SQUARE_2": "4",
				"SQUARE_3": "9",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tmp, err := ioutil.TempDir("", "exports")
			require.NoError(t, err)
			t.Cleanup(func() {
				_ = os.RemoveAll(tmp)
			})

			cmd := exec.Command("bin/exports", fmt.Sprintf("test/fixtures/%s/export", tc.name), ".", tmp)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = tc.environ
			err = cmd.Run()
			require.NoError(t, err)

			gotEnvs := make(map[string]string)
			err = filepath.Walk(tmp, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					if path == tmp {
						return nil
					}

					return fmt.Errorf("unexpected directory %s", path)
				}

				name := strings.SplitN(info.Name(), ".", 2)
				require.Len(t, name, 2)
				assert.Equal(t, "override", name[1])

				value, err := ioutil.ReadFile(path)
				require.NoError(t, err)

				gotEnvs[name[0]] = string(value)

				return nil
			})
			require.NoError(t, err)
			assert.Equal(t, tc.wantEnvs, gotEnvs)
		})
	}
}
