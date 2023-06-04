package dsl

const GrammarDescription = `[DSL Concepts]
A source/sink CLASS is identified by two identifiers separated by a dot ("."), and they are provided by plugins. Both source and sink plugins may require arguments.
(Run 'nomlog plugins' for details)

Certain transformations and all sinks will consume a source. This means that the source IDENTIFIER is no longer valid for consumption.

The general flow of a script is to setup one or more sources, perform any necessary transformations, and output the streams to one or more sinks.
The same thing can be accomplished with Go code, but the DSL syntax is a little more approachable.


[DSL Syntax]
Source identifies a log source and exposes it in the runtime.
  source as IDENTIFIER CLASS [ARG [, ARG]]

Merge will combine two sources with a new identifier. The combined sources will be marked as consumed.
  merge NEW_IDENTIFIER and IDENTIFIER as IDENTIFIER

Dupe will duplicate a source into two new sources. This is useful to output the same log events in two different ways.
The input source will be marked as consumed.
  dupe SRC_IDENTIFIER as NEW_IDENTIFIER and NEW_IDENTIFIER

Append will forward all source stream to a target stream, consuming the source.
  append SRC_IDENTIFIER to TARGET_IDENTIFIER

Cut will will perform an eager field split using a specified delimiter, or the default space.
Sequential delimiters in the log message will all be consumed at once.
  cut [with STRING] IDENTIFIER set(FIELD_IDENTIFIER=FIELD_NUM [, FIELD_IDENTIFIER=FIELD_NUM])

Fanout will spread the log events in one stream into two new streams, consuming the source.
  fanout IDENTIFIER as IDENTIFIER and IDENTIFIER

Tag allows easily attaching string metadata to a stream. The stream will not be consumed.
  tag IDENTIFIER with STRING

Join is useful for combining multi-line, unstructured log output.
Multiple comma-separated regex patterns may be used to specify what makes up a start line.
  join IDENTIFIER with REGEX_STRING [, REGEX_STRING]

Sink writes log entries to a plugin provided output sink. This will consume the specified stream.
  sink IDENTIFIER [async as IDENTIFIER] to CLASS [ARG [, ARG]]
`
