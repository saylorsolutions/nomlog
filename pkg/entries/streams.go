package entries

// Merge will take over the passed in LogIterators and forward all LogEntry elements to the new LogIterator.
// It's advised not to read from an iterator that has been passed to Merge.
func Merge(a, b LogIterator) LogIterator {
	aCh := NewIterationChannel(a)
	bCh := NewIterationChannel(b)

	outCh := make(chan LogEntry)
	out := NewChannelIterator(outCh)

	go func() {
		defer close(outCh)
		for aCh != nil || bCh != nil {
			select {
			case ae, ok := <-aCh:
				if !ok {
					aCh = nil
					continue
				}
				outCh <- ae
			case be, ok := <-bCh:
				if !ok {
					bCh = nil
					continue
				}
				outCh <- be
			}
		}
	}()
	return out
}
