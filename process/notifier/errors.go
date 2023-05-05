package notifier

import "errors"

var errNoSubscribedAddresses = errors.New("no subscribed addresses provided")

var errNoSubscribedIdentifier = errors.New("no subscribed identifier provided")

var errNoSubscribedEvent = errors.New("no subscribed event provided")

var errNilHeaderSubscriber = errors.New("nil header subscriber provided")

var errInvalidHeaderTypeReceived = errors.New("received invalid header type")

var errNilOutportBlock = errors.New("nil outport block provided")

var errNilTransactionPool = errors.New("nil transaction pool provided")

var errNilBlockData = errors.New("nil block data provided")

var errNilHasher = errors.New("nil hasher provided")
