package main

import (
	"crypto/des"
	"encoding/binary"
	"fmt"
	"runtime/debug"
)

func weakhash(key []byte) []byte {
	data := []byte("weakhash")
	n := len(key)
	remainder := n % 8

	var keyLength int
	if remainder != 0 {
		keyLength = n + 8 - remainder
	} else {
		keyLength = n
	}

	passwordKeys := make([]byte, keyLength)
	for i, char := range key {
		passwordKeys[i] = char<<1 + 2
	}

	for i := 0; i < keyLength; i += 8 {
		roundKey := passwordKeys[i : i+8]
		fmt.Println(roundKey)
		cipher, err := des.NewCipher(roundKey)
		if err != nil {
			panic(err)
		}
		cipher.Encrypt(data, data)
	}

	return data
}

func singlehash(plain []byte, key []byte) []byte {
	data := make([]byte, 8)
	copy(data, plain)

	cipher, err := des.NewCipher(key)
	if err != nil {
		panic(err)
	}
	cipher.Encrypt(data, data)

	return data
}

func singleDehash(key []byte) []byte {
	//this line has the real data from the problem
	//data := [2][]byte{[]byte("\xda\x99\xd1\xea\x64\x14\x4f\x3e"), []byte("\x59\xa3\x44\x2d\x8b\xab\xcf\x84")}

	//this is the provided test case
	//data := [1][]byte{[]byte("\xf3\x15\x06\x47\x12\x20\xcd\x8f")}

	//this is my test case with encode by 241 and decode by 123
	data := []byte("\xa5\xc9\x52\xbb\x10\xe3\xd0\x35")

	dehashed := make([]byte, 8)

	cipher, err := des.NewCipher(key)
	if err != nil {
		panic(err)
	}

	cipher.Decrypt(dehashed, data)

	return dehashed
}

func intToString(counter int) []byte {
	bs := make([]byte, 8)
	i := uint64(counter)
	binary.LittleEndian.PutUint64(bs, i)
	return bs
}

func encryptWithAllKeys(start int, stop int, c chan map[string]uint32) {
	r := make(map[string]uint32, 256)
	plain := []byte("weakhash")
	mask := 0x01010101
	counter := start
	i := 0
	for counter < stop {
		counter = (counter | mask) + 1
		key := intToString(counter)
		result := singlehash(plain, key)
		store := string(result)

		r[store] = uint32(counter)
		i++
	}
	c <- r
	fmt.Printf("generated %v hashes from %v to %v \n", i, start, stop)
}

func decryptWithAllKeys(start int, stop int, hashtable map[string]uint32, c chan [2]int) {
	data := []byte("\x59\xa3\x44\x2d\x8b\xab\xcf\x84")
	//data := []byte("\xa5\xc9\x52\xbb\x10\xe3\xd0\x35")
	dehashed := make([]byte, 8)
	counter := start
	i := 0
	//fmt.Printf("calculating from %v to %v\n", counter, stop)
	for counter < stop {
		counter = (counter | 0x0101010101010101) + 1
		key := intToString(counter)
		cipher, err := des.NewCipher(key)
		if err != nil {
			panic(err)
		}

		cipher.Decrypt(dehashed, data)
		store := string(dehashed)

		if k, ok := hashtable[store]; ok {
			fmt.Printf("Key generated from int %d encrypt and int %d found by thread %v in round %v\n", k, counter, start, i)
			res := [2]int{int(k), counter}
			c <- res
			return
		}
		i++
	}
	res := [2]int{0, 0}
	fmt.Printf("looked at %v and nothing found from %v to %v\n last value was %v\n", i, start, stop, counter)

	c <- res

}

func meetInTheMiddle() {

	nHashToGenerate := 1 << 29
	nHashToCheck := 1 << 50
	nThreads := 8
	mask8 := 0x01010101
	mask16 := 0x0101010101010101
	nGenPerThread := nHashToGenerate / nThreads
	nCheckPerThread := nHashToCheck / nThreads
	c := make(chan map[string]uint32, nThreads)

	for i := 0 | mask8; i < nHashToGenerate|mask8; i += nGenPerThread {
		start := i | mask8
		end := start + nGenPerThread
		fmt.Printf("hash gen thread for %v to %v started\n", start, end)
		go encryptWithAllKeys(start, end, c)
	}

	hashtable := make(map[string]uint32, nHashToGenerate)
	for i := 0; i < nThreads; i++ {
		r := <-c
		for k, v := range r {
			hashtable[k] = v
			delete(r, k)
		}
		r = nil
	}

	fmt.Println("generated all hashes")

	output := make(chan [2]int, nThreads)
	//op := len(hashtable)
	for i := 0 | mask16; i < nHashToCheck|mask16; i += nCheckPerThread {
		start := i | mask16
		end := start + nCheckPerThread
		fmt.Printf("hash check thread for %v to %v started\n", start, end)
		go decryptWithAllKeys(start, end, hashtable, output)

	}
	for i := 0; i < nThreads; i++ {
		fmt.Println(<-output)
	}
}

func main() {
	GCPercent := 2
	prev := debug.SetGCPercent(GCPercent)
	fmt.Printf("set GC to %v, prev was %v\n", GCPercent, prev)
	//key := []byte("Hello World")
	/* 	key1 := intToString(241)
	   	key2 := intToString(123)
	   	plain := []byte("weakhash")
	   	fullKey := append(key1, key2...)

	   	fmt.Println(hex.EncodeToString(fullKey))

	   	hashedValue := singlehash(singlehash(plain, key1), key2)
	   	fmt.Println(hex.EncodeToString(hashedValue))
	*/

	meetInTheMiddle()

	/* 	a := 72341272349704449
	   	b := a | 0x0101010101010101
	   	c := b | 0x0101010101010101 + 1
	   	fmt.Println(a)
	   	fmt.Println(b)
	   	fmt.Println(c) */
	/* 	test1 := intToString(16843248)
	   	test2 := intToString(72340172838076794)
	   	fullKey = append(test1, test2...)
	   	fmt.Println(hex.EncodeToString(fullKey))

	   	hashedValue = singlehash(singlehash(plain, test1), test2)
	   	fmt.Println(hex.EncodeToString(hashedValue)) */
}
