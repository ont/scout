package main

import (
	"io"
	"log"
	"regexp"
)

type BagRegexp struct {
	out io.Writer
	r   *regexp.Regexp
}

func NewBagRegexp(reg string, out io.Writer) *BagRegexp {
	return &BagRegexp{
		out: out,
		r:   regexp.MustCompile(reg),
	}
}

func (b *BagRegexp) Write(pair *ReqResPair, bytes []byte) error {
	log.Println(pair.req.Host + pair.req.URL.Path)
	if b.r.MatchString(pair.req.Host + pair.req.URL.Path) {
		_, err := b.out.Write(bytes)
		if err != nil {
			return err
		}

		_, err = b.out.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
