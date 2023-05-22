package dsl

const (
	streamSize = 64
)

type tokenStream struct {
	ch  <-chan token
	buf [streamSize]token
	idx int
	err *token
}

func newTokenStream(ch <-chan token) *tokenStream {
	return &tokenStream{
		ch: ch,
	}
}

func (s *tokenStream) peek() token {
	t := s.next()
	s.pushBack(t)
	return t
}

func (s *tokenStream) next() token {
	if s.err != nil {
		return *s.err
	}
	if s.idx > 0 {
		s.idx--
		return s.buf[s.idx]
	}
	t, ok := <-s.ch
	if !ok {
		return token{Type: tEof}
	}
	if t.Type == tErr {
		s.err = &t
		go func() {
			for range s.ch {
			}
		}()
		return t
	}
	return t
}

func (s *tokenStream) pushBack(tokens ...token) {
	for _, t := range tokens {
		if s.idx == streamSize {
			panic("stream filled to capacity")
		}
		s.buf[s.idx] = t
		s.idx++
	}
}
