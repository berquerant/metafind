package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/berquerant/metafind/expr"
	"github.com/berquerant/metafind/iox"
	"github.com/berquerant/metafind/logx"
	"github.com/berquerant/metafind/meta"
	"github.com/berquerant/metafind/metric"
	"github.com/berquerant/metafind/prober"
	"github.com/berquerant/metafind/walk"
	"github.com/berquerant/metafind/worker"
	"github.com/berquerant/structconfig"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

const (
	bufferSize = 100
)

var (
	errNotSpecified = errors.New("NotSpecified")
	errArgument     = errors.New("Argument")
)

func NewConfig(fs *pflag.FlagSet) (*Config, error) {
	var (
		sc     = NewStructConfig()
		merger = NewConfigMerger()
	)

	if err := sc.SetFlags(fs); err != nil {
		return nil, err
	}
	const configFlag = "config"
	fs.StringP(configFlag, "c", "",
		`Config file.
example:

# root directories (default: [.])
root:
  - ROOT1
# shell command (default: [sh])
sh:
  - bash
probe:
  - ffprobe -v error -hide_banner -show_entries format -of json=c=1 @ARG
expr: |
  name matches '\.m4a$'`)

	if err := fs.Parse(os.Args); err != nil {
		return nil, err
	}

	var (
		parseFile = func(c Config) (Config, error) {
			file, _ := fs.GetString(configFlag)
			if file == "" {
				return c, nil
			}
			var x Config
			if err := sc.FromDefault(&x); err != nil {
				return c, nil
			}
			if err := (&x).parse(file); err != nil {
				return c, err
			}
			return merger.Merge(c, x)
		}
		parseEnv = func(c Config) (Config, error) {
			var x Config
			if err := sc.FromEnv(&x); err != nil {
				return c, err
			}
			return merger.Merge(c, x)
		}
		parseFlag = func(c Config) (Config, error) {
			var x Config
			if err := sc.FromFlags(&x, fs); err != nil {
				return c, err
			}
			return merger.Merge(c, x)
		}
	)

	var config Config
	if err := sc.FromDefault(&config); err != nil {
		return nil, err
	}

	var err error
	for _, f := range []func(Config) (Config, error){
		parseFile,
		parseEnv,
		parseFlag,
	} {
		if config, err = f(config); err != nil {
			return nil, err
		}
	}

	if config.Worker < 1 {
		config.Worker = 1
	}

	return &config, nil
}

func (c *Config) parse(v string) error {
	f, err := os.Open(v)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(b, c); err != nil {
		return err
	}
	return nil
}

type Config struct {
	Debug     bool     `json:"debug" yaml:"debug" name:"debug" usage:"Enable debug logs"`
	Quiet     bool     `json:"quiet" yaml:"quiet" name:"quiet" short:"q" usage:"Quiet logs except ERROR"`
	Verbose   bool     `json:"verbose" yaml:"verbose" name:"verbose" short:"v" usage:"Verbose output. Output metadata to stdout and metrics to stderr"`
	Worker    int      `json:"worker" yaml:"worker" name:"worker" short:"w" default:"8" usage:"Worker num"`
	Out       string   `json:"out" yaml:"out" name:"out" short:"o" usage:"Output file. - means stdout"`
	Root      []string `json:"root" yaml:"root" name:"root" short:"r" default:"." usage:"Root directories. - means stdin; separated by ';'"`
	Shell     []string `json:"shell" yaml:"shell" name:"sh" default:"sh" usage:"Shell command for probe; separated by ';'"`
	Probe     []string `json:"probe" yaml:"probe" name:"probe" short:"p" usage:"Probe script. The script should write json to stdout, called by passing the filepath as the 1st argument. Read script from FILE by '@FILE'; separated by ';'"`
	ProbeName []string `json:"pname" yaml:"pname" name:"pname" usage:"Probe script name. Change metadata name; separated by ';'"`
	Index     []string `json:"index" yaml:"index" name:"index" short:"i" usage:"Read metadata from the specified files instead of scanning the directory. Read metadata from stdin by -; separated by ';'"`
	Expr      string   `json:"expr" yaml:"expr" name:"expr" short:"e" usage:"Expression of expr lang"`
}

func (Config) unmarshalCallback(f structconfig.StructField, v string, fv func() reflect.Value) error {
	n, _ := f.Tag().Name()
	switch n {
	case "root", "sh", "probe", "index", "pname":
		if v == "" {
			return nil
		}
		xs := strings.Split(v, ";")
		fv().Set(reflect.ValueOf(xs))
		return nil
	default:
		return fmt.Errorf("unmarshalCallback: unexpected field: %s=%s", n, v)
	}
}

func (Config) equalCallback(a, b any) (bool, error) {
	switch a := a.(type) {
	case []string:
		b, ok := b.([]string)
		if !ok {
			return false, nil
		}
		return slices.Equal(a, b), nil
	default:
		return false, fmt.Errorf("equalCallback: unexpected type: %#v, %#v", a, b)
	}
}

func NewStructConfig() *structconfig.StructConfig[Config] {
	var c Config
	return structconfig.New[Config](
		structconfig.WithAnyCallback(c.unmarshalCallback),
	)
}

func NewConfigMerger() *structconfig.Merger[Config] {
	var c Config
	return structconfig.NewMerger[Config](
		structconfig.WithAnyCallback(c.unmarshalCallback),
		structconfig.WithAnyEqual(c.equalCallback),
	)
}

func (c Config) NewOutput() (io.WriteCloser, error) {
	if c.Out != "" {
		return iox.NewWriteCloser(os.Stdout, c.Out)
	}
	return iox.AsWriteCloser(os.Stdout), nil
}

func (c Config) NewIndexReader() (iox.ReaderAndCloser, error) {
	if len(c.Index) == 0 {
		return nil, errNotSpecified
	}
	return iox.NewReaderAndCloser(os.Stdin, c.Index...)
}

func (c Config) logLevel() slog.Leveler {
	if c.Debug {
		return slog.LevelDebug
	}
	if c.Quiet {
		return slog.LevelError
	}
	return slog.LevelInfo
}

func (c Config) SetupLogger() {
	logx.Setup(os.Stderr, c.logLevel())
}

func (c *Config) NewExpr() (expr.Expr, error) {
	if c.Expr == "" {
		return nil, errNotSpecified
	}
	code, err := iox.ReadFileOrLiteral(c.Expr)
	if err != nil {
		return nil, err
	}
	return expr.New(code)
}

func (c *Config) newProbers() ([]meta.Prober, error) {
	xs := make([]meta.Prober, len(c.Probe))
	for i, p := range c.Probe {
		code, err := iox.ReadFileOrLiteral(p)
		slog.Debug("newProber", slog.String("p", p), slog.String("code", code), logx.Err(err))
		if err != nil {
			return nil, err
		}
		xs[i] = meta.NewScript(code, c.Shell[0], c.Shell[1:]...)
	}

	return xs, nil
}

func (c *Config) NewRootWalker() (*iox.Walker, error) {
	args := c.Root
	switch {
	case len(args) == 0:
		return nil, fmt.Errorf("%w: no roots", errArgument)
	case !slices.Contains(args, iox.StdinMark):
		w := walk.NewFile()
		return iox.NewWalker(w, args...), nil
	case len(args) == 1:
		w := walk.NewReader(os.Stdin, walk.NewFile())
		return iox.NewWalker(w, args...), nil
	default:
		return nil, fmt.Errorf("%w: no other files can be specifined using %s (stdin)",
			errArgument,
			iox.StdinMark,
		)
	}
}

func (c *Config) NewEntryWorker() *worker.Worker[walk.Entry, *meta.Data] {
	return worker.New(
		"WalkMeta",
		c.Worker,
		func(_ context.Context, x walk.Entry) (*meta.Data, error) {
			return walk.NewMetaData(x), nil
		})
}

func (c *Config) newProberWorkers() ([]*worker.Worker[*meta.Data, *meta.Data], error) {
	probers, err := c.newProbers()
	if err != nil {
		return nil, err
	}
	workerName := func(i int) string {
		if i >= 0 && i < len(c.ProbeName) {
			return c.ProbeName[i]
		}
		return fmt.Sprintf("p%d", i)
	}

	workers := make([]*worker.Worker[*meta.Data, *meta.Data], len(c.Probe))
	for i, p := range probers {
		workers[i] = prober.NewWorker(p, c.Worker, workerName(i))
	}
	return workers, nil
}

func (c *Config) NewProberWorkersChain() (*worker.Chain[*meta.Data], error) {
	workers, err := c.newProberWorkers()
	if err != nil {
		return nil, err
	}
	return worker.NewChain(workers, c.Worker), nil
}

var (
	AcceptCount = metric.NewCounter("Accept")
)

func (c *Config) Output(w io.Writer, v *meta.Data) {
	AcceptCount.Incr()

	if c.Verbose {
		// dump all metadata as json
		b, err := json.Marshal(v)
		if err != nil {
			slog.Warn("Marshal", logx.Err(err))
			return
		}
		fmt.Fprintf(w, "%s\n", b)
		return
	}

	path := walk.GetPathFromMetadata(v)
	fmt.Fprintln(w, path)
}
