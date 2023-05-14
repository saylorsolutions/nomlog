// Package plugin provides functionality that is not available with just iterators and entries.
// Splitting these out into their own, independent (except what's provided in pkg) packages means that they can be omitted in favor of a smaller build size if the functionality isn't needed.
// This likely won't result in shorter initial compile times, since the dependencies are still listed in the root level go.mod.
//
// "Source" functions should take input and return an iterator.Iterator and potentially an error, and operate asynchronously.
// Sources should close any resources, like file handles or channels, and stop the associated goroutine when they have reached the end of their input.
//
// "Sink" functions should take an iterator.Iterator - and optionally other parameters - and operate synchronously (the user may decide to call a Sink function in a goroutine).
// Sink functions should use iterator.Drain on an iterator if they encounter an error to prevent upstream blocking.
//
// A plugin is anything that implements the Plugin interface. A Plugin is expected to register its source and sink functions when its Plugin.Register method is called, and may perform cleanup operations in Plugin.Closing.
// Arbitrary logic may be added around these events as needed. If Plugin.Register is not called, then neither will Plugin.Closing.
//
//	Current Plugins:
//	- file provides source and sink for files, including tail support.
//	- store provides SQLite source and sink.
//
// More will be added as time allows.
package plugin
