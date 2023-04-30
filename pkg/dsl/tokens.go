package dsl

const (
	streamSize = 64
)

type tokenStream struct {
	ch  <-chan token
	buf [streamSize]token
	idx int
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
	if s.idx > 0 {
		s.idx--
		return s.buf[s.idx]
	}
	t, ok := <-s.ch
	if !ok {
		return token{Type: tEof}
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
