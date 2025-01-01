package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/berquerant/metafind/logx"
	"github.com/spf13/pflag"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGPIPE,
	)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("Err", slog.Any("err", err))
	}
}

func run(ctx context.Context) error {
	startTime := time.Now()

	fs := pflag.NewFlagSet("main", pflag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, "mf")
		fs.PrintDefaults()
	}

	config, err := NewConfig(fs)
	if errors.Is(err, pflag.ErrHelp) {
		return nil
	}
	failOnError(err)

	config.SetupLogger()
	slog.Debug("Config", logx.JSON("config", config))
	if config.Verbose {
		defer func() {
			duration := time.Since(startTime)
			fmt.Fprintf(os.Stderr, "%s\n", logx.Jsonify(NewMetrics(duration)))
		}()
	}

	return find(ctx, config)
}

const usage = `%[1]s -- search files by metadata

%[1]s searches for files using metadata.
The search criteria utilizes Expr (https://expr-lang.org/).
If the expression is not specified or if the expression evaluates to true,
the path of that file will be displayed.

The following results are considered true when evaluating the expression:
- boolean true
- non-zero numeric value
- non-empty string, map, or array

Available inputs:
- basename: name but ext
- basepath: path but ext
- dir: All but the last element of path
- ext: The file name extension
- mod_time: The last modification time of the file
- mod_time_ts: The last modification timestamp of the file
- mode: The file permissions (in octal)
- name: The name of the file
- path: The path of the file
- size: The file size (in bytes)

You can add inputs by specifying 'probe'.
The 'probe' is invoked with the path to the target file (1st argument).
You can use the macros within the script.

- @ARG is replaced with "$1"
- @RAWARG is replaced with $1

And the 'probe' must output a JSON string or in the form "key=value" to standard output like:
  {"key": "value"}

  key1=value1
  key2=value2

The keys for the inputs available in the expression will be 'pN' for the N-th 'probe'.

Examples:

# Dump metadata
%[1]s -r SOME_DIR -v
# Search path by regexp
%[1]s -r SOME_DIR -e 'path matches "green"'
# Exclude by expr
%[1]s -r SOME_DIR -x 'path matches "green"'
# Add metadata
%[1]s -r SOME_DIR -e 'p0.path matches "green"' -p 'echo "{\"p\":\"@ARG\"}"'
# Add named metadata
%[1]s -r SOME_DIR -e 'key.path matches "green"' -p 'echo "{\"p\":\"@ARG\"}"' --pname 'key'
# Add equal pair metadata
%[1]s -r SOME_DIR -e 'key.path matches "green"' -p 'echo "p=@RAWARG"' --pname 'key'
# Read paths from stdin
echo SOME_DIR | %[1]s -r - -v
# Read metadata
%[1]s -i METADATA_FILE -e 'path matches "green"'
# Envvars
ROOT=SOME_DIR EXPR='size==0' %[1]s

Flags:

`

func failOnError(err error) {
	if err != nil {
		slog.Error("exit", "err", fmt.Sprintf("%v", err))
		os.Exit(1)
	}
}
