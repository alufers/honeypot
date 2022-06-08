package fakeshell

import "bufio"

type ScanByteCounter struct {
	BytesRead int
}

func (s *ScanByteCounter) Wrap(split bufio.SplitFunc) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		adv, tok, err := split(data, atEOF)
		s.BytesRead += adv
		return adv, tok, err
	}
}
