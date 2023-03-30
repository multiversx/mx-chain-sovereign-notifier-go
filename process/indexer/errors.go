package indexer

import "errors"

var errNilMarshaller = errors.New("nil marshaller provided")

var errNilIndexer = errors.New("nil indexer provided")

var errNilSovereignNotifier = errors.New("nil sovereign notifier provided")

var errOperationTypeInvalid = errors.New("invalid/unknown operation type")
