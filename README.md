# nomlog

This package aims to provide an easy to use, adhoc log management toolkit.
It uses an iterator-based flow for source, sink, interpretation, and transformation.
Transform plain text logging to structured logging with a source and a few iterator operations.
I plan to use compile time plugins to add more distinct, tech-specific functionality, and to support more interesting features later.

## Current Features

* Create log entry iterators from slices or channels.
* Transform log entry fields in flight.
  * Uses generics to provide type safety for transforming any given type that may appear in logs.
* Tag log entry iterators for contextual filtering.
* Cut from log line strings to make more useful structured log entries.
  * This supports both positive and negative field indexing.
* Join log entry messages based on regex pattern(s) that specify what the start of a log entry should look like.
  * Entries that don't match will be joined to a matching start entry (or the first entry in the iterator).
  * This can be useful for catching stack traces and other, less structured information that appears in plain text logs.
* Reassign field values to new field names in flight.
* Merge, duplicate, and split iterators to create more complex data flows.
* Add logic to iterators (like middleware) to filter, cancel, or concatenate them.
* Source and sink from/to files.
* Query from and sink to SQLite (no cgo) using the same iterator pattern.
  * More interesting functionality with SQLite is planned.
* Use a Domain Specific Language (DSL) to describe a log management pipeline.
* Use the nomlog CLI to interact with DSL scripts.
  * Launch a nomlog session from a file with `nomlog exec someFile`.
  * Check that your scripts are valid with `nomlog vet someFile`.

# Contributing to nomlog

I'm happy to accept contributions in terms of documentation, problem or suggestion reporting through GitHub issues, or PRs to add new features/plugins.
It's all helpful!

The idea is to keep this package simple enough that it can be easily used as a library, but powerful enough to support 80% of adhoc log management needs as a CLI.
This can be a hard line to walk, but it's worth it to both solve my own headaches, and hopefully solve that of others too without it becoming a bloated mess.

Stay tuned for updates! 😁

## Next Steps

More functionality is planned as time allows:
* RabbitMQ log sink plugin.
  * This should be an easy win, and it may help to validate performance.
* Remote file tailing over SSH, like what I did in [sshtail](https://github.com/drognisep/sshtail).
* Remote proxying, like local nomlog <-> nomlog on a remote host.
  * This may require scripting, or at least some DSL support.
* Possibly a CUI interface for SQLite files to help digest a log dump that may include multiple tables.
  * This would be really neat to do, but obviously a lot more work than just building the library.
  * I'm looking at either [tview](https://github.com/rivo/tview) or [bubbletea](https://github.com/charmbracelet/bubbletea) as a possibility.
* Basic string argument interpolation to promote reuse and templating.
