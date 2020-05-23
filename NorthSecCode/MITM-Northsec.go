package main

import (
	"crypto/des"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	//"github.com/alecthomas/mph"
	"runtime/debug"
)

var nToMatch int = 8

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

func singleDehash(key, hash []byte) []byte {
	data := hash

	dehashed := make([]byte, 8)

	cipher, err := des.NewCipher(key)
	if err != nil {
		panic(err)
	}

	cipher.Decrypt(dehashed, data)

	return dehashed
}

func intToString(counter uint) []byte {
	bs := make([]byte, 8)
	i := uint64(counter)
	binary.LittleEndian.PutUint64(bs, i)
	return bs
}

func getInt2(s []byte) uint {
	var res uint
	for _, v := range s {
		res <<= 8
		res |= uint(v)
	}
	return res
}

func encryptWithAllKeys(start uint, nToGenerate uint, hashtable map[uint]uint, mutex *sync.Mutex, wg *sync.WaitGroup) {
	plain := []byte("weakhash")
	mask := uint(0x0101010101010101)
	counter := start

	defer wg.Done()

	for i := uint(0); i < nToGenerate; i++ {
		counter = (counter | mask) + 1
		key := intToString(counter)
		result := singlehash(plain, key)
		store := getInt2(result[:nToMatch])

		mutex.Lock()
		hashtable[store] = counter
		mutex.Unlock()
	}

}

func decryptWithAllKeys(start uint, nToGenerate uint, hashtable map[uint]uint, c chan [2]uint) {
	//data := []byte(0x59a3442d8babcf84)
	startTime := time.Now()
	data := [2][]byte{[]byte("\xda\x99\xd1\xea\x64\x14\x4f\x3e"), []byte("\x59\xa3\x44\x2d\x8b\xab\xcf\x84")}

	dehashed := make([]byte, 8)
	mask := uint(0x0101010101010101)
	counter := start

	//fmt.Printf("calculating from %v to %v\n", counter, stop)
	for i := uint(0); i < nToGenerate; i++ {
		counter = (counter | mask) + 1
		key := intToString(counter)
		cipher, err := des.NewCipher(key)
		if err != nil {
			panic(err)
		}
		for _, hash := range data {
			cipher.Decrypt(dehashed, hash)
			store := getInt2(dehashed[:nToMatch])

			if k, ok := hashtable[store]; ok {
				fmt.Printf("Key generated from int 0x%x encrypt and int 0x%x found for hash %x in %v\n", k, counter, hash, time.Now().Sub(startTime))
				res := [2]uint{uint(k), counter}
				c <- res
				return
			}
		}
	}
	res := [2]uint{0, 0}

	c <- res

}

func meetInTheMiddle() {

	nHashToGenerate := uint(1 << 30)
	nHashToCheck := uint(1 << 36)
	nThreads := uint(8)

	nGenPerThread := nHashToGenerate / nThreads
	nCheckPerThread := nHashToCheck / nThreads

	fmt.Println("starting threads")

	increment := uint(1 << 53)
	hashtable := make(map[uint]uint, nHashToGenerate)
	var mutex = &sync.Mutex{}
	var wg sync.WaitGroup

	start := time.Now()

	for i := uint(0); i < nThreads*increment; i += increment {
		start := i | 0x0101010101010101
		wg.Add(1)
		go encryptWithAllKeys(start, nGenPerThread, hashtable, mutex, &wg)
	}
	wg.Wait()
	elapsed := time.Now().Sub(start)
	fmt.Printf("generated %v hashes in %v \n", nHashToGenerate, elapsed)

	output := make(chan [2]uint, nThreads)
	//op := len(hashtable)
	for i := uint(0); i < nThreads*increment; i += increment {
		start := i | 0x0101010101010101

		go decryptWithAllKeys(start, nCheckPerThread, hashtable, output)

	}
	for i := uint(0); i < nThreads; i++ {
		fmt.Println(<-output)
	}
}

func validate(key1, key2 []byte) []byte {
	//key1 := intToString(135390567628899252)
	//key2 := intToString(126383387400464336)
	plain := []byte("weakhash")

	fullKey := append(key1, key2...)
	fmt.Println(hex.EncodeToString(fullKey))

	weakhashedValue := singlehash(singlehash(plain, key1), key2)
	fmt.Println(hex.EncodeToString(weakhashedValue))

	dehashed := singleDehash(key2, []byte("\x59\xa3\x44\x2d\x8b\xab\xcf\x84"))
	fmt.Println(hex.EncodeToString(dehashed))

	hashed := singlehash(plain, key1)
	fmt.Println(hex.EncodeToString(hashed))

	return hashed
}

func main() {
	GCPercent := 5
	prev := debug.SetGCPercent(GCPercent)
	fmt.Printf("set GC to %d, prev was %d\n", GCPercent, prev)
	//key := []byte("Hello World")

	meetInTheMiddle()

}
