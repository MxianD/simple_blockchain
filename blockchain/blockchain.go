package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type BlockChain struct {
	NodeIdentifier      string //chain-id
	Chain               []Block
	Nodes               []string
	CurrentTransactions []Transaction //info that hasn't be packaged
}

type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"receipient"`
	Amount    uint   `json:"amount"`
}

type Block struct {
	Index       int           `json:"index"`
	TimeStamp   int64         `json:"timeStamp"`
	Transaction []Transaction `json:"transactions"`
	Pow         uint64        `json:"pow"`
	PrevHash    string        `json:"prevHash"`
}

// constructor
func (b *BlockChain) Init(nodeIdentifier string) {
	b.Nodes = []string{}

	firstBlock := Block{
		Index:       0,
		TimeStamp:   time.Now().Unix(),
		Transaction: nil,
		Pow:         1,
		PrevHash:    "1",
	}
	b.Chain = append(b.Chain, firstBlock)
}

func (b *BlockChain) RegisterNode(node string) {
	b.Nodes = append(b.Nodes, node)
}

func (b *BlockChain) GetNodes() []string {
	return b.Nodes
}

func (b *BlockChain) GetBlockHash(block Block) string {
	blockBytes, _ := json.Marshal(block)
	prevHashValueTemp := sha256.Sum256(blockBytes)
	return hex.EncodeToString(prevHashValueTemp[:])
}

func (b *BlockChain) VerifyChain(chain []Block) bool {
	if len(chain) <= 1 {
		return true
	}
	lastBlock := chain[0]
	curChainIdx := 1

	for curChainIdx < len(chain) {
		curBlock := chain[curChainIdx]

		if curBlock.PrevHash != b.GetBlockHash(lastBlock) {
			return false
		}

		if !b.VerifyPOW(lastBlock.Pow, curBlock.Pow) {
			return false
		}
		lastBlock = curBlock
		curChainIdx++
	}
	return true
}

type FullChainResp struct {
	Chain        []Block       `json:"chain"`
	Len          int           `json:"len"`
	Transactions []Transaction `json:"transactions"`
	Nodes        []string      `json:"nodes"`
}

// resolve conflicts
func (b *BlockChain) ResolveConflicts() {

	var maxLenChain []Block
	var maxLen = len(b.Chain)

	// send request to neighbour nodes, and verify their chain
	for _, node := range b.Nodes {
		resp, err := http.Get(node + "/fullChain") //send http request to acquire the full chain
		if err != nil {
			fmt.Println(err)
			continue
		}

		defer resp.Body.Close()

		fullChainRespBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Failed to read response body:", err)
			return
		}

		var fullChainResp FullChainResp
		err = json.Unmarshal(fullChainRespBytes, &fullChainResp)
		if err != nil {
			fmt.Println("Failed to read response body:", err)
			return
		}

		if len(fullChainResp.Chain) != fullChainResp.Len {
			continue
		}

		if fullChainResp.Len > maxLen && b.VerifyChain(fullChainResp.Chain) {
			maxLen = fullChainResp.Len
			maxLenChain = fullChainResp.Chain
		}

	}

	// if find a longer chain, update
	if maxLen > len(b.Chain) {
		b.Chain = maxLenChain
	}

}

// add a new transaction, need to be mined into a new block
func (b *BlockChain) NewTransaction(newTransaction Transaction) {
	b.CurrentTransactions = append(b.CurrentTransactions, newTransaction)
}

func (b *BlockChain) GetCurrentTransactionsSize() int {
	return len(b.CurrentTransactions)
}

func (b *BlockChain) NewBlock(proof uint64) Block {
	prevHash := b.GetBlockHash(b.Chain[len(b.Chain)-1])

	block := Block{
		Index:       len(b.Chain),
		TimeStamp:   time.Now().Unix(),
		Transaction: b.CurrentTransactions,
		Pow:         proof,
		PrevHash:    prevHash,
	}

	b.Chain = append(b.Chain, block)
	b.CurrentTransactions = []Transaction{}

	return block
}

// verify a proof of work
func (b BlockChain) VerifyPOW(lastProof uint64, curProof uint64) bool {
	lastProofStr := strconv.FormatUint(lastProof, 10)
	curProofStr := strconv.FormatUint(curProof, 10)
	hasValueTmp := sha256.Sum256([]byte(lastProofStr + curProofStr))
	hasValue := hex.EncodeToString(hasValueTmp[:])

	if hasValue[0:2] == "00" {
		return true
	} else {
		return false
	}
}

func (b *BlockChain) GetNodeIdentifier() string {
	return b.NodeIdentifier
}

func (b *BlockChain) GetPOW() uint64 {
	lastProof := b.Chain[len(b.Chain)-1].Pow

	var curProof uint64 = 0

	for !b.VerifyPOW(lastProof, curProof) {
		curProof++
	}

	return curProof
}
