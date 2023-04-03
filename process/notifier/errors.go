package notifier

import "errors"

var errNoSubscribedAddresses = errors.New("no subscribed addresses provided")

var errDuplicateSubscribedAddresses = errors.New("duplicate subscribed addresses provided")

var errNilExtendedHeaderHandler = errors.New("nil extended handler provided")

var errReceivedHeaderType = errors.New("received invalid header type")

var errNilOutportblock = errors.New("nil outport block provided")

var errNilTransactionPool = errors.New("nil transaction pool provided")

var errNilBlockData = errors.New("nil block data provided")
