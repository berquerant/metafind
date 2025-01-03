# metafind

``` shell
❯ mf --help
mf -- search files by metadata

mf searches for files using metadata.
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
mf -r SOME_DIR -v
# Search path by regexp
mf -r SOME_DIR -e 'path matches "green"'
# Exclude by expr
mf -r SOME_DIR -x 'path matches "green"'
# Add metadata
mf -r SOME_DIR -e 'p0.p matches "green"' -p 'echo "{\"p\":\"@ARG\"}"'
# Add named metadata
mf -r SOME_DIR -e 'key.p matches "green"' -p 'echo "{\"p\":\"@ARG\"}"' --pname 'key'
# Add equal pair metadata
mf -r SOME_DIR -e 'key.p matches "green"' -p 'echo "p=@RAWARG"' --pname 'key'
# Read paths from stdin
echo SOME_DIR | mf -r - -v
# Read metadata
mf -i METADATA_FILE -e 'path matches "green"'
# Envvars
ROOT=SOME_DIR EXPR='size==0' mf
# Format by expr
mf -r SOME_DIR -f '{n:name,s:size}'

Flags:

  -c, --config string    Config file.
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
                           name matches '\.m4a$'
      --debug            Enable debug logs
  -x, --exclude string   Expression of expr lang to reject entries before probe
  -e, --expr string      Expression of expr lang to select entries
  -f, --format string    Expression of expr lang to format output
  -i, --index string     Read metadata from the specified files instead of scanning the directory. Read metadata from stdin by -; separated by ';'
  -o, --out string       Output file. - means stdout
      --pname string     Probe script name. Change metadata name; separated by ';'
  -p, --probe string     Probe script. The script should write json to stdout, called by passing the filepath as the 1st argument. Read script from FILE by '@FILE'; separated by '#'
  -q, --quiet            Quiet logs except ERROR
  -r, --root string      Root directories. - means stdin; separated by ';' (default ".")
      --sh string        Shell command for probe; separated by ';' (default "sh")
  -v, --verbose          Verbose output. Output metadata to stdout and metrics to stderr
  -w, --worker int       Worker num (default 8)
```
