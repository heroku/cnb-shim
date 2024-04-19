package cnbshim_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	tmp, err := ioutil.TempDir("", "build")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(tmp)
	})

	// set up test env
	require.NoError(t, os.MkdirAll(tmp+"/layers", 0755))
	require.NoError(t, os.MkdirAll(tmp+"/platform", 0755))
	require.NoError(t, os.MkdirAll(tmp+"/buildpack/target", 0755))
	require.NoError(t, os.MkdirAll(tmp+"/buildpack/bin", 0755))

	require.NoError(t, exec.Command("cp", "-r", "test/fixtures/build/app", tmp+"/").Run())
	require.NoError(t, exec.Command("cp", "-r", "test/fixtures/build/buildpack/.", tmp+"/buildpack/target/").Run())
	require.NoError(t, exec.Command("cp", "bin/build", tmp+"/buildpack/bin/").Run())
	// add fake exports and release binaries
	require.NoError(t, ioutil.WriteFile(tmp+"/buildpack/bin/exports", nil, 0755))
	require.NoError(t, ioutil.WriteFile(tmp+"/buildpack/bin/release", nil, 0755))

	// run bin/build
	var out bytes.Buffer
	cmd := exec.Command(tmp+"/buildpack/bin/build", tmp+"/layers", tmp+"/platform")
	cmd.Dir = tmp + "/app"
	cmd.Env = append(os.Environ(), "CNB_STACK_ID=heroku-20", "ALLOW_EOL_SHIMMED_BUILDER=1")
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if _, ok := err.(*exec.ExitError); err != nil && ok {
		t.Logf("bin/build output:\n%s", out.String())
	}
	require.NoError(t, err)

	contains := []string{
		"got STACK=heroku-20",
		fmt.Sprintf("got arg 0=%s/app", tmp),
		fmt.Sprintf("got arg 1=%s/layers/shim", tmp),
		fmt.Sprintf("got arg 2=%s/platform/env", tmp),
	}
	for _, c := range contains {
		assert.Contains(t, out.String(), c)
	}

	files := []string{
		"/layers/profile.toml",
		"/layers/profile/env.build",
		"/layers/profile/profile.d/1.sh",
	}
	for _, f := range files {
		_, err := os.Stat(tmp + f)
		assert.NoError(t, err, f)
	}

	out = bytes.Buffer{}
	cmd = exec.Command("bash", "-c", fmt.Sprintf(`
echo
echo "before HOME=$HOME"
source "%s"
echo "after HOME=$HOME"
`, tmp+"/layers/profile/profile.d/1.sh"))
	cmd.Dir = tmp + "/app"
	cmd.Env = []string{"HOME=/home/app"}
	cmd.Stdout = &out
	cmd.Stderr = &out
	require.NoError(t, cmd.Run())
	assert.Equal(t, fmt.Sprintf(`
before HOME=/home/app
buildpack HOME=%s
after HOME=/home/app
`, tmp+"/app"), out.String())
}
