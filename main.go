package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	math "math/rand"
	"time"
)

type Block struct {
	Timestamp     string
	PrevHash      string
	Hash          string
	ValidatorAddr string
}

func (n PoSNetwork) PrintBlockchainInfo() {
	for i, block := range n.Blockchain {
		fmt.Printf("Block %d Info:\n", i)
		block.PrintBlockInfo()
	}
}

func (b Block) PrintBlockInfo() {
	fmt.Println("\tTimestamp:", b.Timestamp)
	fmt.Println("\tPrevious Hash:", b.PrevHash)
	fmt.Println("\tHash:", b.Hash)
	fmt.Println("\tValidator Address:", b.ValidatorAddr)
}

type PoSNetwork struct {
	Blockchain     []*Block
	BlockchainHead *Block
	Validators     []*Node
}

type Node struct {
	Stake   int
	Address string
}

func (n PoSNetwork) GenerateNewBlock(Validator *Node) ([]*Block, *Block, error) {
	if err := n.ValidateBlockchain(); err != nil {
		Validator.Stake -= 10
		return n.Blockchain, n.BlockchainHead, err
	}

	currentTime := time.Now().String()

	newBlock := &Block{
		Timestamp:     currentTime,
		PrevHash:      n.BlockchainHead.Hash,
		Hash:          NewBlockHash(n.BlockchainHead),
		ValidatorAddr: Validator.Address,
	}

	if err := n.ValidateBlockCandidate(newBlock); err != nil {
		Validator.Stake -= 10
		return n.Blockchain, n.BlockchainHead, err
	} else {
		n.Blockchain = append(n.Blockchain, newBlock)
	}
	return n.Blockchain, newBlock, nil
}

func NewBlockHash(block *Block) string {
	blockInfo := block.Timestamp + block.PrevHash + block.Hash + block.ValidatorAddr
	return newHash(blockInfo)
}

func newHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func (n PoSNetwork) ValidateBlockchain() error {
	if len(n.Blockchain) <= 1 {
		return nil
	}

	currBlockIdx := len(n.Blockchain) - 1
	prevBlockIdx := len(n.Blockchain) - 2

	for prevBlockIdx >= 0 {
		currBlock := n.Blockchain[currBlockIdx]
		prevBlock := n.Blockchain[prevBlockIdx]
		if currBlock.PrevHash != prevBlock.Hash {
			return errors.New("blockchain has inconsistent hashes")
		}

		if currBlock.Timestamp <= prevBlock.Timestamp {
			return errors.New("blockchain has inconsistent timestamps")
		}

		if NewBlockHash(prevBlock) != currBlock.Hash {
			return errors.New("blockchain has inconsistent hash generation")
		}
		currBlockIdx--
		prevBlockIdx--
	}
	return nil
}
func (n PoSNetwork) NewNode(stake int) []*Node {
	newNode := &Node{
		Stake:   stake,
		Address: randAddress(),
	}
	n.Validators = append(n.Validators, newNode)
	return n.Validators
}

func randAddress() string {
	b := make([]byte, 16)
	_, _ = math.Read(b)
	return fmt.Sprintf("%x", b)
}

func (n PoSNetwork) SelectWinner() (*Node, error) {
	var winnerPool []*Node
	totalStake := 0
	for _, node := range n.Validators {
		if node.Stake > 0 {
			winnerPool = append(winnerPool, node)
			totalStake += node.Stake
		}
	}
	if winnerPool == nil {
		return nil, errors.New("there are no nodes with stake in the network")
	}
	winnerNumber := math.Intn(totalStake)
	tmp := 0
	for _, node := range n.Validators {
		tmp += node.Stake
		if winnerNumber < tmp {
			return node, nil
		}
	}
	return nil, errors.New("a winner should have been picked but wasn't")
}
func (n PoSNetwork) ValidateBlockCandidate(newBlock *Block) error {
	if n.BlockchainHead.Hash != newBlock.PrevHash {
		return errors.New("blockchain HEAD hash is not equal to new block previous hash")
	}

	if n.BlockchainHead.Timestamp >= newBlock.Timestamp {
		return errors.New("blockchain HEAD timestamp is greater than or equal to new block timestamp")
	}

	if NewBlockHash(n.BlockchainHead) != newBlock.Hash {
		return errors.New("new block hash of blockchain HEAD does not equal new block hash")
	}
	return nil
}

func main() {
	// set random seed
	math.Seed(time.Now().UnixNano())

	// generate an initial PoS network including a blockchain with a genesis block.
	genesisTime := time.Now().String()
	pos := &PoSNetwork{
		Blockchain: []*Block{
			{
				Timestamp:     genesisTime,
				PrevHash:      "",
				Hash:          newHash(genesisTime),
				ValidatorAddr: "",
			},
		},
	}
	pos.BlockchainHead = pos.Blockchain[0]

	// instantiate nodes to act as validators in our network
	pos.Validators = pos.NewNode(60)
	pos.Validators = pos.NewNode(40)

	// build 5 additions to the blockchain
	for i := 0; i < 4; i++ {
		winner, err := pos.SelectWinner()
		if err != nil {
			log.Fatal(err)
		}
		winner.Stake += 10
		pos.Blockchain, pos.BlockchainHead, err = pos.GenerateNewBlock(winner)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Round ", i)
		fmt.Println("\tAddress:", pos.Validators[0].Address, "-Stake:", pos.Validators[0].Stake)
		fmt.Println("\tAddress:", pos.Validators[1].Address, "-Stake:", pos.Validators[1].Stake)
	}

	pos.PrintBlockchainInfo()

}
