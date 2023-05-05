package factory

import "errors"

var errNoSubscribedAddresses = errors.New("no subscribed addresses provided")

var errDuplicateSubscribedAddresses = errors.New("duplicate subscribed addresses provided")
