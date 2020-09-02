package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/common/log"
)

// NewBlockChain : Returns new instantiated Blockchain struct
func NewBlockChain() *BlockChain {
	blockChain := &BlockChain{}
	blockChain.chain = make([]*Block, 0)
	blockChain.currentTransactions = make([]*Transaction, 0)
	blockChain.nodes = make([]string, 0)

	seedHash := [32]byte{}
	blockChain.newBlock("1", hex.EncodeToString(seedHash[:])) // seed block

	return blockChain
}

// BlockChain : The block chain object, containing the transactions, nodes, and chain
type BlockChain struct {
	chain               []*Block
	currentTransactions []*Transaction
	nodes               []string
}

func (b *BlockChain) newBlock(proof string, previousHash string) *Block {

	block := &Block{
		Index:        len(b.chain) + 1,
		Timestamp:    time.Now(),
		Transactions: b.currentTransactions,
		Proof:        proof,
		PreviousHash: previousHash,
	}

	// Reset current Transactions
	b.currentTransactions = make([]*Transaction, 0)
	b.chain = append(b.chain, block)

	log.Info(fmt.Sprintf("New block added"))

	return block
}

func (b *BlockChain) newTransaction(sender string, recipient string, amount int) int {
	b.currentTransactions = append(b.currentTransactions, &Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	})
	log.Info(fmt.Sprintf("Added new transaction"))
	return b.lastBlock().Index + 1
}

func (b *BlockChain) hash(block *Block) string {
	marshalledJSON, _ := json.Marshal(*block)
	hash := sha256.Sum256(marshalledJSON)
	return hex.EncodeToString(hash[:])
}

func (b *BlockChain) lastBlock() *Block {
	return b.chain[len(b.chain)-1]
}

func (b *BlockChain) proofOfWork(lastProof string) string {
	proof := 0
	for !b.validProof(lastProof, string(proof)) {
		proof++
	}
	return string(proof)
}

func (b *BlockChain) validProof(lastProof string, proof string) bool {
	guessHash := sha256.Sum256([]byte(lastProof + proof))
	guesshashString := hex.EncodeToString(guessHash[:])
	lastFour := guesshashString[len(guesshashString)-1 : len(guesshashString)]
	return string(lastFour) == "0"
}

func (b *BlockChain) validChain(chain []*Block) bool {
	lastblock := chain[0]
	currentIndex := 1

	for currentIndex < len(chain) {
		block := chain[currentIndex]

		if block.PreviousHash != b.hash(lastblock) {
			return false
		}

		if !b.validProof(lastblock.Proof, block.Proof) {
			return false
		}

		lastblock = block
		currentIndex++
	}

	return true
}

func (b *BlockChain) resolveConfilcts() bool {
	maxLength := len(b.chain)
	var newChain []*Block

	client := &http.Client{Timeout: 10 * time.Second}

	for _, node := range b.nodes {
		resp, err := client.Get(fmt.Sprintf("%s/chain", node))
		if err != nil {
			log.Error("Unable to retrieve node")
		}
		defer resp.Body.Close()
		var respBody ChainDTO
		json.NewDecoder(resp.Body).Decode(&respBody)
		if respBody.Length > maxLength && b.validChain(respBody.Chain) {
			maxLength = respBody.Length
			newChain = respBody.Chain
		}
	}

	if newChain != nil {
		b.chain = newChain
		return true
	}
	return false
}

func (b *BlockChain) registerNode(node string) {
	if !containsNode(b.nodes, node) {
		b.nodes = append(b.nodes, node)
		log.Info(fmt.Sprintf("Registered node  %s", node))
	}
}

func containsNode(nodes []string, targetNode string) bool {
	for _, node := range nodes {
		if node == targetNode {
			return true
		}
	}
	return false
}
