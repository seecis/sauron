package extractor

import "bufio"

type Field struct {
	label     string
	data      string
	subfields []Field
}

type Extractor interface {
	Extract(reader *bufio.Reader) (*Field, error)
}
