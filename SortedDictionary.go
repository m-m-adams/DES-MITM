package main

import (
	"fmt"
	"sort"
)

/* HMap makes a sorted array based on the weakhash. This is slower to construct than
a standard Go map, but uses massively less memory and has comparable lookups.
Note that it is impossible to add values to this map, you have to remake it.
Map creation currently takes more memory than necessary, it's on my todo list*/

//HMap is a "map" based on using the crypto hash as the lookup. Sort by value than lookup with binary search
type HMap struct {
	NKeys  int
	Keys   []uint
	Values []uint
}

//NewHMap creats an MPHmap from matching key and entry slices
func NewHMap(keys, values []uint) *HMap {
	var hmap HMap
	hmap.NKeys = int(len(keys))
	hmap.Keys = keys
	hmap.Values = values

	sort.Sort(&hmap)
	fmt.Println("finished sorting")
	return &hmap
}

//Lookup the value matching the key
func (hmap *HMap) Lookup(key uint) (uint, bool) {

	index := sort.Search(hmap.NKeys, func(i int) bool { return hmap.Keys[i] >= key })

	if index < hmap.NKeys && hmap.Keys[index] == key {
		return hmap.Values[index], true
	}
	return 0, false

}

func (hmap *HMap) Len() int {
	return hmap.NKeys
}

func (hmap *HMap) Less(i, j int) bool {

	return hmap.Keys[i] < hmap.Keys[j]

}

func (hmap *HMap) Swap(i, j int) {
	hmap.Keys[i], hmap.Keys[j] = hmap.Keys[j], hmap.Keys[i]
	hmap.Values[i], hmap.Values[j] = hmap.Values[j], hmap.Values[i]
}

/* func main() {
	nTest := uint(1 << 28)
	keys := make([]uint, nTest)
	values := make([]uint, nTest)

	testMap := make(map[uint]uint, nTest)

	for i := range keys {
		x := rand.uint()
		y := rand.uint()
		keys[i] = x
		values[i] = y
		testMap[x] = y
	}
	hmap := NewHMap(keys, values)

	fmt.Println("running the test")

	for _, k := range keys {

		hmapValue, _ := hmap.Lookup(k)

		if hmapValue != testMap[k] {
			fmt.Println("no match")
		}
	}
	fmt.Println("done")
}
*/
