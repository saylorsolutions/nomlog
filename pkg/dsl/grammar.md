# DSL Grammar
This describes the grammar associated with nomlog files.

## Literals

```
EOL        := "\n"
STRING     := """ '[^"]*' """
NUMBER     := '-?\d+(\.\d+)?'
LPAR       := "("
RPAR       := ")"
INT        := '-?\d+'
EQ         := "="
COMMA      := ","
AS         := "as"
AND        := "and"
TO         := "to"
SOURCE     := "source"
VAR        := "var"
SINK       := "sink"
ASYNC      := "async"
IDENTIFIER := '\w[\w\d]*'
MERGE      := "merge"
DUPE       := "dupe"
APPEND     := "append"
CUT        := "cut"
SET        := "set
WITH       := "with"
FANOUT     := "fanout"
TAG        := "tag"
CLASS      := '\w+\.\w+'
```

## Productions
There are some dynamically defined literals used in these productions.
* **source_class:** defines a type of source, like `file.Tail`.
* **sink_class:** defines a type of sink, like `file.Sink`.
* **arg:** A dynamically defined value that is specific to the `source_class` or `sink_class` that precedes it.

```
eol          := (EOL|EOF)
arg          := (STRING|NUMBER|INT|IDENTIFIER)
args         := arg (COMMA arg)*
source_class := IDENTIFIER DOT IDENTIFIER
source       := SOURCE AS IDENTIFIER source_class args eol
sink_class   := IDENTIFIER DOT IDENTIFIER
sink         := SINK IDENTIFIER TO sink_class args eol
async_sink   := SINK IDENTIFIER ASYNC AS IDENTIFIER TO sink_class args eol
merge        := MERGE IDENTIFIER AND IDENTIFIER AS IDENTIFIER eol
dupe         := DUPE IDENTIFIER AS IDENTIFIER AND IDENTIFIER eol
append       := APPEND IDENTIFIER TO IDENTIFIER eol
cut          := CUT (WITH STRING)? IDENTIFIER SET LPAR IDENTIFIER EQ INT ("," IDENTIFIER EQ INT)* RPAR eol
fanout       := FANOUT IDENTIFIER AS IDENTIFIER AND IDENTIFIER eol
tag          := TAG IDENTIFIER WITH STRING eol
```
