package main

import (
	"regexp"
)

type BagRegexp struct {
	out Storage
	r   *regexp.Regexp
}

func NewBagRegexp(reg string, out Storage) *BagRegexp {
	return &BagRegexp{
		out: out,
		r:   regexp.MustCompile(reg),
	}
}

func (b *BagRegexp) Write(pair *ReqResPair) error {
	//log.Println(pair.req.Host + pair.req.URL.Path)
	if pair.req == nil {
		return nil // skip pair without request
	}

	if b.r.MatchString(pair.req.Host + pair.req.URL.Path) {
		err := b.out.Save(pair)
		if err != nil {
			return err
		}
	}
	return nil
}
