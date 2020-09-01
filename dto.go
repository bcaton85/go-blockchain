package main

type ChainDTO struct {
	Chain  []*Block
	Length int
}

type NodeDTO struct {
	Nodes []string
}
