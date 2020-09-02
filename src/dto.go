package blockchain

type ChainDTO struct {
	Chain  []*Block
	Length int
}

type NodeDTO struct {
	Nodes []string
}
