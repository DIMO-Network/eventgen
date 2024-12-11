// Pacakge event contains the CloudEvent wrapper structs for on-chain events.
package event

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const eventType = "zone.dimo.contract.event"

type Block struct {
	Number *big.Int    `json:"number,omitempty"`
	Hash   common.Hash `json:"hash,omitempty"`
	Time   time.Time   `json:"time,omitempty"`
}

type LogInfo struct {
	ChainID         int64          `json:"chainId"`
	EventName       string         `json:"eventName"`
	Block           Block          `json:"block,omitempty"`
	Contract        common.Address `json:"contract"`
	TransactionHash common.Hash    `json:"transactionHash"`
	EventSignature  common.Hash    `json:"eventSignature"`
}

type Data[A any] struct {
	LogInfo
	Arguments A `json:"arguments"`
}
