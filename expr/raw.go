package expr

import (
	"fmt"
	"log/slog"

	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/metric"
	exprl "github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type RawExpr interface {
	Run(env map[string]any) (any, error)
}

var (
	_ RawExpr = &RawProgram{}
)

type RawProgram struct {
	program *vm.Program
}

func MustNewRaw(code string) *RawProgram {
	p, err := NewRaw(code)
	if err != nil {
		panic(err)
	}
	return p
}

func NewRaw(code string) (*RawProgram, error) {
	p, err := exprl.Compile(code)
	slog.Debug("NewRawExpr", slog.String("code", code), logx.Err(err))
	if err != nil {
		return nil, err
	}
	return &RawProgram{
		program: p,
	}, nil
}

var (
	RawRunCount = metric.NewCounter("RawExprRun")
	RawErrCount = metric.NewCounter("RawExprErr")
)

func (p *RawProgram) Run(env map[string]any) (any, error) {
	RawRunCount.Incr()

	v, err := exprl.Run(p.program, env)
	slog.Debug("RawExprRun",
		logx.JSON("env", env),
		slog.String("return", fmt.Sprint(v)),
		logx.Err(err),
	)
	if err != nil {
		RawErrCount.Incr()
		return nil, err
	}

	return v, nil
}
