package notifier

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/config"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/process"
)

var log = logger.GetOrCreate("notifier-sovereign-process")

type ExtendedHeaderHandler func(header *block.ShardHeaderExtended)

type sovereignNotifier struct {
	mutHandler          sync.RWMutex
	handlers            []process.ExtendedHeaderHandler
	subscribedAddresses [][]byte
	blockContainer      process.BlockContainerHandler
	marshaller          marshal.Marshalizer
}

// NewSovereignNotifier will create a sovereign shard notifier
func NewSovereignNotifier(config config.Config, blockContainer process.BlockContainerHandler, marshaller marshal.Marshalizer) (*sovereignNotifier, error) {
	addresses := config.SubscribedAddresses
	if len(addresses) == 0 {
		return nil, errNoSubscribedAddresses
	}

	log.Debug("received config", "subscribed addresses", addresses)

	return &sovereignNotifier{
		subscribedAddresses: nil, // todo here use bytes addresses
		handlers:            make([]process.ExtendedHeaderHandler, 0),
		mutHandler:          sync.RWMutex{},
		blockContainer:      blockContainer,
		marshaller:          marshaller,
	}, nil
}

// Notify will notify the sovereign nodes via p2p about the finalized block and incoming mb txs
func (notifier *sovereignNotifier) Notify(outportBlock *outport.OutportBlock) error {
	mbs, err := notifier.getAllIncomingMbs(outportBlock.TransactionPool)
	if err != nil {
		return err
	}

	headerV2, err := notifier.getHeaderV2(core.HeaderType(outportBlock.BlockData.HeaderType), outportBlock.BlockData.HeaderBytes)
	if err != nil {
		return err
	}

	extendedHeader := &block.ShardHeaderExtended{
		Header:             headerV2,
		IncomingMiniBlocks: mbs,
	}

	notifier.notifyHandlers(extendedHeader)
	return nil
}

func (notifier *sovereignNotifier) getHeaderV2(headerType core.HeaderType, headerBytes []byte) (header *block.HeaderV2, err error) {
	creator, err := notifier.blockContainer.Get(headerType)
	if err != nil {
		return nil, err
	}

	headerHandler, err := block.GetHeaderFromBytes(notifier.marshaller, creator, headerBytes)
	if err != nil {
		return nil, err
	}

	headerV2, castOk := headerHandler.(*block.HeaderV2)
	if !castOk {
		return nil, fmt.Errorf("invalid")
	}

	return headerV2, nil
}

func (notifier *sovereignNotifier) notifyHandlers(extendedHeader *block.ShardHeaderExtended) {
	notifier.mutHandler.RLock()
	defer notifier.mutHandler.RUnlock()

	for _, handler := range notifier.handlers {
		go handler.SaveExtendedHeader(extendedHeader)
	}
}

func (notifier *sovereignNotifier) getAllIncomingMbs(txPool *outport.TransactionPool) ([]*block.MiniBlock, error) {
	mbs := make([]*block.MiniBlock, 0)

	txsMb, err := notifier.getIncomingMbFromTxs(txPool.Transactions)
	if err != nil {
		return nil, err
	}

	// TODO: when specs are defined, we should also handle scrs mbs
	// so we would also append smth like scrMb here
	mbs = append(mbs, txsMb)
	return mbs, nil
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

func (notifier *sovereignNotifier) RegisterHandler(handler process.ExtendedHeaderHandler) error {
	if handler == nil {
		return fmt.Errorf("invalid")
	}

	notifier.mutHandler.Lock()
	notifier.handlers = append(notifier.handlers, handler)
	notifier.mutHandler.Unlock()

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (notifier *sovereignNotifier) IsInterfaceNil() bool {
	return notifier == nil
}
