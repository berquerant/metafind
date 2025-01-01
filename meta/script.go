package meta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/berquerant/execx"
	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/metric"
)

type Prober interface {
	Probe(ctx context.Context, path string) (*Data, error)
}

var _ Prober = &Script{}

type Script struct {
	s *execx.Script
}

const (
	ArgLiteral    = "@ARG"
	RawArgLiteral = "@RAWARG"
)

func ReplaceScriptLiterals(s string) string {
	r := strings.NewReplacer(
		RawArgLiteral, `$1`,
		ArgLiteral, `"$1"`,
	)
	return r.Replace(s)
}

func NewScript(content, shell string, arg ...string) *Script {
	slog.Debug("NewScript",
		slog.String("shell", shell),
		slog.String("arg", strings.Join(arg, " ")),
		slog.String("content", content),
	)

	ScriptCount.Incr()
	content = ReplaceScriptLiterals(content)
	s := execx.NewScript(content, shell, arg...)
	s.KeepScriptFile = true
	s.Env.Merge(execx.EnvFromEnviron())
	return &Script{
		s: s,
	}
}

var (
	ScriptCount       = metric.NewCounter("MetaScript")
	ProbeCount        = metric.NewCounter("MetaProbe")
	ProbeSuccessCount = metric.NewCounter("MetaProbeSuccess")
	ProbeFailureCount = metric.NewCounter("MetaProbeFailure")
)

var (
	ErrParse = errors.New("Parse")
)

func (s *Script) Probe(ctx context.Context, path string) (*Data, error) {
	ProbeCount.Incr()
	var data *Data

	if err := s.s.Runner(func(cmd *execx.Cmd) error {
		cmd.Args = append(cmd.Args, path)
		r, err := cmd.Run(ctx)
		if err != nil {
			return fmt.Errorf("%w: cmd.run: args=%s", err, logx.Jsonify(cmd.Args))
		}

		b, err := io.ReadAll(r.Stdout)
		if err != nil {
			return fmt.Errorf("%w: read stdout", err)
		}

		data = s.parseData(b)
		if data.IsEmpty() {
			return fmt.Errorf("%w: %s", ErrParse, b)
		}
		return nil
	}); err != nil {
		ProbeFailureCount.Incr()
		return nil, err
	}

	ProbeSuccessCount.Incr()
	return data, nil
}

func (Script) parseData(b []byte) *Data {
	d := map[string]any{}
	if err := json.Unmarshal(b, &d); err == nil {
		return NewData(d)
	}
	return NewDataFromEqualPairs(strings.Split(string(b), "\n"))
}

func (s *Script) Close() error {
	return s.s.Close()
}
