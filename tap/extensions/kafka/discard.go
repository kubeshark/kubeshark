package kafka

import "bufio"

func discardN(r *bufio.Reader, sz int, n int) (int, error) {
	var err error
	if n <= sz {
		n, err = r.Discard(n)
	} else {
		n, err = r.Discard(sz)
		if err == nil {
			err = errShortRead
		}
	}
	return sz - n, err
}
