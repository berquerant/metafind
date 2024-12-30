package expr

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/metric"
	exprl "github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type Expr interface {
	Run(env map[string]any) (bool, error)
}

var (
	_ Expr = &Program{}
)

type Program struct {
	program *vm.Program
}

func New(code string) (*Program, error) {
	p, err := exprl.Compile(code)
	slog.Debug("NewExpr", slog.String("code", code), logx.Err(err))
	if err != nil {
		return nil, err
	}
	return &Program{
		program: p,
	}, nil
}

var (
	RunCount   = metric.NewCounter("ExprRun")
	ErrCount   = metric.NewCounter("ExprErr")
	TrueCount  = metric.NewCounter("ExprTrue")
	FalseCount = metric.NewCounter("ExprFalse")
)

func (p *Program) Run(env map[string]any) (bool, error) {
	RunCount.Incr()

	v, err := exprl.Run(p.program, env)
	slog.Debug("ExprRun",
		logx.JSON("env", env),
		slog.String("return", fmt.Sprint(v)),
		logx.Err(err),
	)
	if err != nil {
		ErrCount.Incr()
		return false, err
	}
	if AsBool(v) {
		TrueCount.Incr()
		return true, nil
	}
	FalseCount.Incr()
	return false, nil
}

func AsBool(v any) bool {
	if v == nil {
		return false
	}

	tv := reflect.ValueOf(v)
	switch reflect.TypeOf(v).Kind() {
	case reflect.Bool:
		return tv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return tv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return tv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return tv.Float() != 0
	case reflect.String:
		return tv.String() != ""
	case reflect.Map, reflect.Array, reflect.Slice:
		return tv.Len() > 0
	default:
		return false
	}
}
