package notifier

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

type txsNotifier struct {
	mutSubscribers sync.RWMutex
	subscribers    []process.TransactionSubscriber
}

func newTxsNotifier() *txsNotifier {
	return &txsNotifier{
		mutSubscribers: sync.RWMutex{},
		subscribers:    make([]process.TransactionSubscriber, 0),
	}
}

func (tn *txsNotifier) registerSubscriber(handler process.TransactionSubscriber) error {
	if check.IfNil(handler) {
		return errNilTxSubscriber
	}

	tn.mutSubscribers.Lock()
	tn.subscribers = append(tn.subscribers, handler)
	tn.mutSubscribers.Unlock()

	return nil
}

func (tn *txsNotifier) notifyTxSubscribers(txs []*txHandlerInfo) {
	tn.mutSubscribers.RLock()
	defer tn.mutSubscribers.RUnlock()

	log.Info("notifying incoming txs", "num txs", len(txs))

	for _, subscriber := range tn.subscribers {
		notifyTxSubscriber(subscriber, txs)
	}
}

func notifyTxSubscriber(subscriber process.TransactionSubscriber, txs []*txHandlerInfo) {
	// TODO: Use real sender shard ID once sovereign shard can accept it (nodes coordinator refactor)
	//cacheID := fmt.Sprintf("%d_%d", txInfo.senderShardID, core.SovereignChainShardId)
	cacheID := fmt.Sprintf("%d_%d", core.MainChainShardId, core.SovereignChainShardId)
	for _, txInfo := range txs {
		subscriber.AddData(txInfo.hash, txInfo.tx, txInfo.tx.Size(), cacheID)
	}
}
