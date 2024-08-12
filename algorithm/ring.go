package algorithm

import (
	"errors"
	"github.com/liangyaopei/consistent"
)

/*
 * consistent node ring face
 * - base on `github.com/liangyaopei/consistent`
 */

//face info
type Ring struct {
	hashRing *consistent.HashRing
}

//construct
func NewRing() *Ring {
	this := &Ring{
		hashRing: consistent.New(nil, consistent.DefaultHashFn),
	}
	return this
}

//get all ring nodes
func (f *Ring) GetAllNodes() []string {
	//call base func
	nodes := f.hashRing.GetNodes()
	return nodes
}

//get all node weight
func (f *Ring) GetAllNodeWeight() map[string]int {
	//call base func
	nodesWeight := f.hashRing.GetNodeWeight()
	return nodesWeight
}

//get key node
func (f *Ring) GetNode(key string) (string, error) {
	//check
	if key == "" {
		return "", errors.New("invalid parameter")
	}

	//call base func
	node := f.hashRing.LocateKeyStr(key)
	return node, nil
}

//add nodes
func (f *Ring) AddNodes(nodes ...string) error {
	//check
	if nodes == nil || len(nodes) <= 0 {
		return errors.New("invalid parameter")
	}

	//call base func
	for _, node := range nodes {
		f.hashRing.AddNode(node)
	}
	return nil
}

//add one node with weight
func (f *Ring) AddNodeWithWeight(node string, weight int) error {
	//check
	if node == "" || weight <= 0 {
		return errors.New("invalid parameter")
	}

	//call base func
	f.hashRing.AddNodeWeight(node, weight)
	return nil
}

//update node with weight
func (f *Ring) UpdateNode(node string, weight int) error {
	//check
	if node == "" || weight < 0 {
		return errors.New("invalid parameter")
	}

	//call base func
	f.hashRing.UpdateNodeWeight(node, weight)
	return nil
}

//del node
func (f *Ring) DelNode(node string) error {
	//check
	if node == "" {
		return errors.New("invalid parameter")
	}

	//call base func
	f.hashRing.DelNode(node)
	return nil
}