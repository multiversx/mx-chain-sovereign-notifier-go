package notifier

import "errors"

var errNilShardCoordinator = errors.New("nil shard coordinator provided")

var errNoSubscribedAddresses = errors.New("no subscribed addresses provided")

var errDuplicateSubscribedAddresses = errors.New("duplicate subscribed addresses provided")

var errNilHeaderSubscriber = errors.New("nil header subscriber provided")

var errNilTxSubscriber = errors.New("nil transaction subscriber provided")

var errInvalidHeaderTypeReceived = errors.New("received invalid header type")

var errNilOutportBlock = errors.New("nil outport block provided")

var errNilTransactionPool = errors.New("nil transaction pool provided")

var errNilBlockData = errors.New("nil block data provided")

var errNilHasher = errors.New("nil hasher provided")
