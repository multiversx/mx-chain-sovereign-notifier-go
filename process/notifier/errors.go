package notifier

import "errors"

var errNoSubscribedAddresses = errors.New("no subscribed addresses provided")

var errNilBlockContainer = errors.New("nil block container provided")

var errNilExtendedHeaderHandler = errors.New("nil extended handler provided")

var errReceivedHeaderType = errors.New("received invalid header type")

var errCannotCastToHeaderV2 = errors.New("cannot cast header to headerV2")

var errNilOutportblock = errors.New("nil outport block provided")

var errNilTransactionPool = errors.New("nil transaction pool provided")

var errNilBlockData = errors.New("nil block data provided")
