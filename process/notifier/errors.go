package notifier

import "errors"

var errNoSubscribedAddresses = errors.New("no subscribed addresses provided")

var errDuplicateSubscribedAddresses = errors.New("duplicate subscribed addresses provided")

var errNilExtendedHeaderHandler = errors.New("nil extended header handler provided")

var errInvalidHeaderTypeReceived = errors.New("received invalid header type")

var errNilOutportBlock = errors.New("nil outport block provided")

var errNilTransactionPool = errors.New("nil transaction pool provided")

var errNilBlockData = errors.New("nil block data provided")

var errNilHasher = errors.New("nil hasher provided")
