package indexer

import "errors"

var errNilMarshaller = errors.New("nil marshaller provided")

var errNilIndexer = errors.New("nil indexer provided")

var errNilSovereignNotifier = errors.New("nil sovereign notifier provided")

var errNilOutportBlockCache = errors.New("nil outport block cache provided")
