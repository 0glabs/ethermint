package types

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
)

//go:generate go run github.com/fjl/gencodec -type Header -field-override headerMarshaling -out gen_header_json.go
//go:generate go run ../../rlp/rlpgen -type Header -out gen_header_rlp.go

// Header represents a block header in the Ethereum blockchain.
type EthHeader struct {
	ParentHash  common.Hash          `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash          `json:"sha3Uncles"       gencodec:"required"`
	Coinbase    common.Address       `json:"miner"`
	Root        common.Hash          `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash          `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash          `json:"receiptsRoot"     gencodec:"required"`
	Bloom       gethtypes.Bloom      `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int             `json:"difficulty"       gencodec:"required"`
	Number      *big.Int             `json:"number"           gencodec:"required"`
	GasLimit    uint64               `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64               `json:"gasUsed"          gencodec:"required"`
	Time        uint64               `json:"timestamp"        gencodec:"required"`
	Extra       []byte               `json:"extraData"        gencodec:"required"`
	MixDigest   common.Hash          `json:"mixHash"`
	Nonce       gethtypes.BlockNonce `json:"nonce"`

	// BaseFee was added by EIP-1559 and is ignored in legacy headers.
	BaseFee *big.Int `json:"baseFeePerGas""`

	/*
		TODO (MariusVanDerWijden) Add this field once needed
		// Random was added during the merge and contains the BeaconState randomness
		Random common.Hash `json:"random" rlp:"optional"`
	*/
	CosmosHeaderHash common.Hash
}

// field type overrides for gencodec
type headerMarshaling struct {
	Difficulty *hexutil.Big
	Number     *hexutil.Big
	GasLimit   hexutil.Uint64
	GasUsed    hexutil.Uint64
	Time       hexutil.Uint64
	Extra      hexutil.Bytes
	BaseFee    *hexutil.Big
	Hash       common.Hash `json:"hash"` // adds call to Hash() in MarshalJSON
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *EthHeader) Hash() common.Hash {
	// replace with cometbft block hash in order to the user who subscribed
	// the "newHeads" message by web socket can get block by "eth_getBlockByHash" successfully.
	return h.CosmosHeaderHash
}

var headerSize = common.StorageSize(reflect.TypeOf(EthHeader{}).Size())

// Size returns the approximate memory used by all internal contents. It is used
// to approximate and limit the memory consumption of various caches.
func (h *EthHeader) Size() common.StorageSize {
	return headerSize + common.StorageSize(len(h.Extra)+(h.Difficulty.BitLen()+h.Number.BitLen())/8)
}

// SanityCheck checks a few basic things -- these checks are way beyond what
// any 'sane' production values should hold, and can mainly be used to prevent
// that the unbounded fields are stuffed with junk data to add processing
// overhead
func (h *EthHeader) SanityCheck() error {
	if h.Number != nil && !h.Number.IsUint64() {
		return fmt.Errorf("too large block number: bitlen %d", h.Number.BitLen())
	}
	if h.Difficulty != nil {
		if diffLen := h.Difficulty.BitLen(); diffLen > 80 {
			return fmt.Errorf("too large block difficulty: bitlen %d", diffLen)
		}
	}
	if eLen := len(h.Extra); eLen > 100*1024 {
		return fmt.Errorf("too large block extradata: size %d", eLen)
	}
	if h.BaseFee != nil {
		if bfLen := h.BaseFee.BitLen(); bfLen > 256 {
			return fmt.Errorf("too large base fee: bitlen %d", bfLen)
		}
	}
	return nil
}

// EmptyBody returns true if there is no additional 'body' to complete the header
// that is: no transactions and no uncles.
func (h *EthHeader) EmptyBody() bool {
	return h.TxHash == gethtypes.EmptyRootHash && h.UncleHash == gethtypes.EmptyUncleHash
}

// EmptyReceipts returns true if there are no receipts for this header/block.
func (h *EthHeader) EmptyReceipts() bool {
	return h.ReceiptHash == gethtypes.EmptyRootHash
}
