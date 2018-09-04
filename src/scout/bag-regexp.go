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
	if !pair.HasRequest {
		return nil // skip pair without request part
	}

	if b.r.MatchString(pair.Host + pair.Path) {
		err := b.out.Save(pair)
		if err != nil {
			return err
		}
	}
	return nil
}
