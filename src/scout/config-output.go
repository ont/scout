package main

import (
	"log"
	"strings"
	"time"
)

func ConfiguredStorage(basename, indexBy string) Storage {
	if indexBy != "" {
		return NewStorageIndexed(basename, indexBy)
	} else {
		return NewStorageBase(basename)
	}
}

func ConfiguredStorageRotate(basename, indexBy string, duration time.Duration) *StorageRotate {
	storage := NewRotateStorage(func(stamp string) Storage {
		return ConfiguredStorage(basename+"-"+stamp, indexBy)
	})
	go storage.RotateEvery(duration)
	return storage
}

func ConfiguredDumper(basename, indexBy string, bags []string, duration time.Duration) *Dumper {
	if basename == "" {
		return NewDumper(nil)
	}

	dumper := NewDumper(ConfiguredStorageRotate(basename, indexBy, duration))

	for _, bag := range bags {
		parts := strings.Split(bag, "::")

		if len(parts) < 2 {
			log.Fatal("Wrong format for bag, must be in form 'regexp::filename'")
		}

		dumper.AddBag(NewBagRegexp(parts[0], ConfiguredStorageRotate(basename, indexBy, duration)))
	}

	return dumper
}
