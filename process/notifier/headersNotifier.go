package notifier

import (
	"encoding/hex"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type headersNotifier struct {
	mutSubscribers sync.RWMutex
	subscribers    []process.HeaderSubscriber
	marshaller     marshal.Marshalizer
	hasher         hashing.Hasher
}

func newHeadersNotifier(marshaller marshal.Marshalizer, hasher hashing.Hasher) *headersNotifier {
	return &headersNotifier{
		mutSubscribers: sync.RWMutex{},
		subscribers:    make([]process.HeaderSubscriber, 0),
		marshaller:     marshaller,
		hasher:         hasher,
	}
}

func (hn *headersNotifier) registerSubscriber(handler process.HeaderSubscriber) error {
	if check.IfNil(handler) {
		return errNilHeaderSubscriber
	}

	hn.mutSubscribers.Lock()
	hn.subscribers = append(hn.subscribers, handler)
	hn.mutSubscribers.Unlock()

	return nil
}

func (hn *headersNotifier) notifyHeaderSubscribers(header data.HeaderHandler) error {
	headerHash, err := core.CalculateHash(hn.marshaller, hn.hasher, header)
	if err != nil {
		return err
	}

	log.Info("notifying shard extended header", "hash", hex.EncodeToString(headerHash))

	hn.mutSubscribers.RLock()
	defer hn.mutSubscribers.RUnlock()

	for _, handler := range hn.subscribers {
		handler.AddHeader(headerHash, header)
	}

	return nil
}
