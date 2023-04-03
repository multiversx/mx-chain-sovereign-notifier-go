package indexer

import "errors"

var errNilMarshaller = errors.New("nil marshaller provided")

var errNilIndexer = errors.New("nil indexer provided")

var errNilSovereignNotifier = errors.New("nil sovereign notifier provided")

var errOperationTypeInvalid = errors.New("invalid/unknown operation type")

var errNilOutportBlockCache = errors.New("nil outport block cache provided")

var errOutportBlockNotFound = errors.New("outport block not found in cache")

var errNilOutportBlock = errors.New("nil outport block provided to be added in cache")

var errOutportBlockAlreadyExists = errors.New("outport block already exists in cache")
