package blockchain

import "time"

type Block struct {
	Index        int
	Timestamp    time.Time
	Transactions []*Transaction
	Proof        string
	PreviousHash string
}
