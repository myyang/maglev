// Package maglev provides implementation of Maglev
package maglev

import (
	"fmt"
	"hash/crc64"
)

const (
	offsetKey = 0xabcaaefe
	skipKey   = 0xa0129efe
)

func crc64fn(src string, key uint64) uint64 {
	return crc64.Checksum([]byte(src), crc64.MakeTable(key))
}

// NewMaglev return maglev by given info
func NewMaglev(nodes []string, m uint64) *Maglev {
	return NewCustomMaglev(nodes, m, crc64fn)
}

// NewCustomMaglev return maglev by more customizable info
func NewCustomMaglev(nodes []string, m uint64, fn HashFunc) *Maglev {
	ml := &Maglev{n: uint64(len(nodes)), nodes: nodes, m: m, fn: fn}
	ml.generatePopulation()
	ml.populate()
	return ml
}

type maglevError struct {
	Message string
}

func (e maglevError) Error() string {
	return fmt.Sprintf("maglevError: %v", e.Message)
}

// HashFunc defines hash function
type HashFunc func(src string, key uint64) uint64

// Maglev struct
type Maglev struct {
	n           uint64 // nodes count
	m           uint64 // lookup table size, MUST BE PRIME NUMBER!!!!!
	permutation [][]uint64
	lookup      []int64
	nodes       []string
	fn          HashFunc
}

func (ml *Maglev) generatePopulation() {
	ml.permutation = make([][]uint64, len(ml.nodes))
	for k, v := range ml.nodes {
		offset := ml.fn(v, offsetKey) % ml.m
		skip := ml.fn(v, skipKey)%(ml.m-1) + 1
		r := make([]uint64, ml.m)
		var j uint64
		for j = 0; j < ml.m; j++ {
			r[j] = (offset + uint64(j)*skip) % ml.m
		}
		ml.permutation[k] = r
	}
}

func (ml *Maglev) populate() {
	next := make([]uint64, ml.n)
	entry := make([]int64, ml.m)
	var i uint64
	for i = 0; i < ml.m; i++ {
		entry[i] = -1
	}
	n := uint64(0)
	for {
		var i uint64
		for i = 0; i < ml.n; i++ {
			c := ml.permutation[i][next[i]]
			for entry[c] >= 0 {
				next[i] = next[i] + 1
				c = ml.permutation[i][next[i]]
			}

			entry[c] = int64(i)
			next[i]++
			n++
			if n == ml.m {
				ml.lookup = entry
				return
			}
		}
	}
}

// AddNode to existing maglev cluster
func (ml *Maglev) AddNode(node string) error {
	for i := 0; i < len(ml.nodes); i++ {
		if ml.nodes[i] == node {
			return maglevError{Message: "Node Exists"}
		}
	}

	ml.nodes = append(ml.nodes, node)
	ml.n = uint64(len(ml.nodes))
	ml.generatePopulation()
	ml.populate()
	return nil
}

// RemoveNode from existing maglev cluster
func (ml *Maglev) RemoveNode(node string) error {
	for i := 0; i < len(ml.nodes); i++ {
		if ml.nodes[i] == node {
			ml.nodes = append(ml.nodes[:i], ml.nodes[i+1:]...)
			ml.n = uint64(len(ml.nodes))
			ml.generatePopulation()
			ml.populate()
			return nil
		}
	}
	return maglevError{Message: "No such exists"}
}

// Get backend through host
func (ml *Maglev) Get(key string) (string, error) {
	if len(ml.nodes) == 0 {
		return "", maglevError{"Empty nodes"}
	}
	i := ml.fn(key, offsetKey)
	return ml.nodes[ml.lookup[i%ml.m]], nil
}
