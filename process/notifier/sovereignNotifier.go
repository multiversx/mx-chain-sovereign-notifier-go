package notifier

import (
	"bytes"
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
)

var log = logger.GetOrCreate("notifier-sovereign-process")

type sovereignNotifier struct {
	subscribedAddresses [][]byte
}

// NewSovereignNotifier will create a sovereign shard notifier
func NewSovereignNotifier(config config.Config) (*sovereignNotifier, error) {
	addresses := config.SubscribedAddresses
	if len(addresses) == 0 {
		return nil, errNoSubscribedAddresses
	}

	log.Debug("received config", "subscribed addresses", addresses)

	return &sovereignNotifier{
		subscribedAddresses: nil, // todo here use bytes addresses
	}, nil
}

// Notify will notify the sovereign nodes via p2p about the finalized block and incoming mb txs
func (notifier *sovereignNotifier) Notify(outportBlock *outport.OutportBlock) error {
	mbs := make([]*block.MiniBlock, 0)

	txsMb, err := notifier.getIncomingMbFromTxs(outportBlock.TransactionPool.Transactions)
	if err != nil {
		return err
	}
	// TODO: when specs are defined, we should also handle scrs mbs
	// Here we will notify all registered handlers about incoming mbs in the future PR
	mbs = append(mbs, txsMb)
	return nil
}

func (notifier *sovereignNotifier) getIncomingMbFromTxs(txs map[string]*outport.TxInfo) (*block.MiniBlock, error) {
	txHashes := make([][]byte, 0)

	for txHash, tx := range txs {
		if !contains(notifier.subscribedAddresses, tx.GetTransaction().GetRcvAddr()) {
			continue
		}

		hashBytes, err := hex.DecodeString(txHash)
		if err != nil {
			return nil, err
		}

		txHashes = append(txHashes, hashBytes)
	}

	return &block.MiniBlock{
		TxHashes:        txHashes,
		ReceiverShardID: 0, // todo: decide what we should fill here
		SenderShardID:   0, // todo: decide what we should fill here
		Type:            block.TxBlock,
		Reserved:        nil,
	}, nil
}

func contains(addresses [][]byte, address []byte) bool {
	for _, addr := range addresses {
		if bytes.Equal(address, addr) {
			return true
		}
	}

	return false
}

// IsInterfaceNil checks if the underlying pointer is nil
func (notifier *sovereignNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
