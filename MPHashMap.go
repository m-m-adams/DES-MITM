package main

import (
	"fmt"
	"sort"
)

/* HMap implements a MPH based hash map. This is slower to construct than
a standard Go map, but uses massively less memory and has faster lookups.
Note that it is impossible to add values to this map, you have to remake it. */

//HMap is a "map" based on using the crypto hash as the lookup. Sort by value than lookup with binary search
type HMap struct {
	NKeys   int
	KVPairs [][2]uint64
}

//NewHMap creats an MPHmap from matching key and entry slices
func NewHMap(kvps [2][]uint64) *HMap {
	var hmap HMap
	hmap.NKeys = len(kvps[0])
	hmap.KVPairs = make([][2]uint64, hmap.NKeys)

	for i := range kvps[0] {
		hmap.KVPairs[i] = [2]uint64{kvps[0][i], kvps[1][i]}

	}
	fmt.Println("finished placing keys")
	hmap.sortValues()
	fmt.Println("finished sorting")
	return &hmap
}

//Lookup the value matching the key
func (hmap *HMap) Lookup(key uint64) (uint64, bool) {

	index := sort.Search(hmap.NKeys, func(i int) bool { return hmap.KVPairs[i][0] >= key })

	if index < hmap.NKeys && hmap.KVPairs[index][0] == key {
		return hmap.KVPairs[index][1], true
	}
	return 0, false

}

func (hmap *HMap) sortValues() {
	sort.Slice(hmap.KVPairs, func(i, j int) bool {
		return hmap.KVPairs[i][0] < hmap.KVPairs[j][0]
	})

}

func main() {
	x := []uint64{700, 100, 600, 200, 300, 400, 500}
	y := []uint64{70, 10, 60, 20, 30, 40, 50}

	xy := [2][]uint64{x, y}

	test := []uint64{100, 543, 121, 200, 300, 1, 400, 500}

	hmap := NewHMap(xy)

	for _, key := range test {
		value, _ := hmap.Lookup(key)
		fmt.Printf("key %v has value %v\n", key, value)
	}
	fmt.Println(hmap.KVPairs)
}
