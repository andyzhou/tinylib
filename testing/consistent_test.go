package testing

import (
	"fmt"
	"github.com/liangyaopei/consistent"
	"testing"
)

func TestConsistent(t *testing.T) {
	nodes := []string{
		"185.199.110.153",
		"185.199.110.154",
		"185.199.110.155",
	}
	ring := consistent.New(nodes, consistent.DefaultHashFn)

	keys := []string{"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	//nodeAdd := "185.199.110.156"
	//nodeDel := "185.199.110.153"

	keys = make([]string, 0)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%v", i)
		keys = append(keys, key)
	}

	//oriDis := make(map[string]string)
	for _, key := range keys {
		node := ring.LocateKeyStr(key)
		t.Logf("key:%v, node:%v\n", key, node)
		//oriDis[key] = node
	}

	//add new node
	nodeAdd := "185.199.110.156"
	ring.AddNodeWeight(nodeAdd, 1)

	t.Logf("====2====\n")
	for _, key := range keys {
		node := ring.LocateKeyStr(key)
		t.Logf("key1:%v, node:%v\n", key, node)
		//oriDis[key] = node
	}

	//ring.AddNodeWeight(nodeAdd, 1)
	//addDes := make(map[string]string)
	//for _, key := range keys {
	//	node := ring.LocateKeyStr(key)
	//	addDes[key] = node
	//}
	//
	//ring.DelNode(nodeDel)
	//delDes := make(map[string]string)
	//for _, key := range keys {
	//	node := ring.LocateKeyStr(key)
	//	delDes[key] = node
	//}
	//
	//t.Logf("adding node:%s,del node:%s", nodeAdd, nodeDel)
	//
	//for _, key := range keys {
	//	t.Logf("key:%15s,ori:%15s,add:%15s,del:%15s", key, oriDis[key], addDes[key], delDes[key])
	//}
	//
	//for node, weight := range ring.GetNodeWeight() {
	//	t.Logf("node:%s,weight:%d", node, weight)
	//}
	//
	//for node, keys := range ring.GetHashRing() {
	//	t.Logf("node:%s,keys:%v", node, keys)
	//}
}