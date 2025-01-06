package main_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	e := newExecutor(t)
	defer e.close()

	t.Run("help", func(t *testing.T) {
		_, err := run(nil, nil, e.cmd, "--help")
		assert.Nil(t, err)
	})

	var (
		d    = filepath.Join(t.TempDir(), "data")
		join = func(s ...string) string {
			return filepath.Join(append([]string{d}, s...)...)
		}
		// trim = func(s string) string {
		// 	return strings.TrimLeft(s, d)
		// }
		newFile = func(name, content string) string {
			p := join(name)
			f, err := os.Create(p)
			if err != nil {
				t.Error(err)
			}
			fmt.Fprint(f, content)
			f.Close()
			return p
		}

		eqWant = func(t *testing.T, x, y []string) {
			slices.Sort(x)
			slices.Sort(y)
			x = slices.DeleteFunc(x, func(a string) bool { return a == "" })
			y = slices.DeleteFunc(y, func(a string) bool { return a == "" })
			assert.Equal(t, x, y)
		}
	)

	if err := os.MkdirAll(d, 0755); err != nil {
		t.Error(err)
	}
	var (
		f1 = newFile("green", "GREEN")
		f2 = newFile("red", "RED")
		f3 = newFile("green2", "GREEN2")
		f4 = newFile("empty", "")
		s1 = newFile("script", `cat <<EOS
{
  "p": @ARG
}
EOS`)
	)

	for _, tc := range []struct {
		title string
		stdin io.Reader
		args  []string
		want  []string
	}{
		{
			title: "exclude",
			args: []string{
				"-r", d,
				"-x", `name == 'script'`,
			},
			want: []string{
				f1,
				f2,
				f3,
				f4,
			},
		},
		{
			title: "equal pairs probe",
			args: []string{
				"-r", d,
				"-e", `c.p matches 'green$'`,
				"-p", `echo "p=@RAWARG"`,
				"--pname", "c",
			},
			want: []string{
				f1,
			},
		},
		{
			title: "named probe",
			args: []string{
				"-r", d,
				"-e", `c.p matches 'green$'`,
				"-p", `echo "{\"p\":\"@ARG\"}"`,
				"--pname", "c",
			},
			want: []string{
				f1,
			},
		},
		{
			title: "probe script",
			args: []string{
				"-r", d,
				"-e", `p0.p matches 'green$'`,
				"-p", "@" + s1,
			},
			want: []string{
				f1,
			},
		},
		{
			title: "probe",
			args: []string{
				"-r", d,
				"-e", `p0.p matches 'green$'`,
				"-p", `echo "{\"p\":\"@ARG\"}"`,
			},
			want: []string{
				f1,
			},
		},
		{
			title: "name is red",
			args: []string{
				"-r", d,
				"-e", `name == 'red'`,
			},
			want: []string{
				f2,
			},
		},
		{
			title: "all paths from stdin",
			stdin: bytes.NewBufferString(d),
			args: []string{
				"-r", "-",
			},
			want: []string{
				f1,
				f2,
				f3,
				f4,
				s1,
			},
		},
		{
			title: "all paths",
			args: []string{
				"-r", d,
			},
			want: []string{
				f1,
				f2,
				f3,
				f4,
				s1,
			},
		},
		{
			title: "format",
			args: []string{
				"-r", d,
				"-f", `name`,
			},
			want: []string{
				`"green"`,
				`"red"`,
				`"green2"`,
				`"empty"`,
				`"script"`,
			},
		},
		{
			title: "format wins verbose",
			args: []string{
				"-r", d,
				"-f", `name`,
				"-v",
			},
			want: []string{
				`"green"`,
				`"red"`,
				`"green2"`,
				`"empty"`,
				`"script"`,
			},
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			got, err := run(tc.stdin, nil, e.cmd, tc.args...)
			if !assert.Nil(t, err) {
				return
			}
			ss := strings.Split(string(got), "\n")
			eqWant(t, tc.want, ss)
		})
	}

	t.Run("env", func(t *testing.T) {
		env := []string{
			fmt.Sprintf("ROOT=%s", d),
			`EXPR=size == 0`,
		}
		got, err := run(nil, env, e.cmd)
		assert.Nil(t, err)
		ss := strings.Split(string(got), "\n")
		eqWant(t, []string{f4}, ss)
	})

	t.Run("index", func(t *testing.T) {
		index, err := run(nil, nil, e.cmd, "-r", d, "-v")
		assert.Nil(t, err)
		stdin := bytes.NewBuffer(index)
		got, err := run(stdin, nil, e.cmd, "-i", "-", "-e", `name == "green2"`)
		assert.Nil(t, err)
		ss := strings.Split(string(got), "\n")
		eqWant(t, []string{f3}, ss)
	})

	t.Run("config", func(t *testing.T) {
		c := newFile("config", fmt.Sprintf(`root:
  - "%s"
probe:
  - |
    echo "{\"p\":\"@ARG\"}"
expr: |
  name contains "green"`, d))
		got, err := run(nil, nil, e.cmd, "-c", c)
		assert.Nil(t, err)
		ss := strings.Split(string(got), "\n")
		eqWant(t, []string{f1, f3}, ss)
	})
}

func run(
	stdin io.Reader,
	env []string,
	name string,
	arg ...string,
) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	cmd.Env = env
	cmd.Dir = "."
	if stdin != nil {
		cmd.Stdin = stdin
	}
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

type executor struct {
	dir string
	cmd string
}

func newExecutor(t *testing.T) *executor {
	t.Helper()
	e := &executor{}
	e.init(t)
	return e
}

func (e *executor) init(t *testing.T) {
	t.Helper()
	const cmdName = "mf"
	dir, err := os.MkdirTemp("", cmdName)
	if err != nil {
		t.Fatal(err)
	}
	cmd := filepath.Join(dir, cmdName)
	// build gbrowse command
	if _, err := run(nil, nil, "go", "build", "-o", cmd); err != nil {
		t.Fatal(err)
	}
	e.dir = dir
	e.cmd = cmd
}

func (e *executor) close() {
	os.RemoveAll(e.dir)
}
