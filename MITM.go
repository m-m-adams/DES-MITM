package main

import (
	"crypto/des"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

var nBytesToMatch int = 8

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

func encryptWithAllKeys(start uint, encryptionKeys []uint, hashes []uint, wg *sync.WaitGroup) {
	plain := []byte("weakhash")
	mask := uint(0x0101010101010101)
	counter := start
	nToGen := uint(len(encryptionKeys))

	defer wg.Done()

	for i := uint(0); i < nToGen; i++ {
		counter = (counter | mask) + 1
		key := intToString(counter)
		result := singlehash(plain, key)
		store := getInt2(result[:nBytesToMatch])

		//keys refers to encryption keys. The output hash is the dictionary key. This is needlessly confusing
		encryptionKeys[i] = counter
		hashes[i] = store
		//fmt.Printf("%x, %x, %x\n", keys[i], getInt2(result), values[i])

	}
	//fmt.Printf("generated %d hashes from %x to %x \n", nToGen, start, counter)
}

func decryptWithAllKeys(start uint, nToGenerate uint, hashtable *HMap, c chan [2]uint) {
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
			store := getInt2(dehashed[:nBytesToMatch])

			if k, ok := hashtable.Lookup(store); ok {
				fmt.Printf("Key generated from int 0x%x encrypt and int 0x%x found for hash %x in %v\n", k, counter, hash, time.Now().Sub(startTime))
				fmt.Printf("%x from %x\n", store, dehashed)
				res := [2]uint{k, counter}
				c <- res
				return
			}
		}
	}
	res := [2]uint{0, 0}

	c <- res

}

func meetInTheMiddle() {

	nHashToGenerate := uint(1 << 31)
	nHashToCheck := uint(1 << 36)
	nThreads := uint(8)

	nGenPerThread := nHashToGenerate / nThreads
	nCheckPerThread := nHashToCheck / nThreads

	var wg sync.WaitGroup

	fmt.Println("starting threads")

	increment := uint(1 << 53)
	encryptkeys := make([]uint, int(nHashToGenerate))
	fmt.Printf("allocated space for %v operations\n", len(encryptkeys))
	hashvalues := make([]uint, int(nHashToGenerate))
	arrayStart := uint(0)

	fmt.Printf("generating hashes\n")
	start := time.Now()
	for i := uint(0); i < nThreads*increment; i += increment {
		wg.Add(1)
		start := i | 0x0101010101010101
		arrayEnd := arrayStart + nGenPerThread

		//fmt.Printf("hash gen thread for %x to %x started\n", start, start+8*nGenPerThread)
		go encryptWithAllKeys(start, encryptkeys[arrayStart:arrayEnd], hashvalues[arrayStart:arrayEnd], &wg)
		//fmt.Println(arrayStart, arrayEnd)
		arrayStart = arrayEnd
	}
	wg.Wait()
	doneHashing := time.Now()
	fmt.Printf("generated %v hashes in %v\n", nHashToGenerate, doneHashing.Sub(start))

	fmt.Println("generating dictionary")
	hashtable := NewHMap(hashvalues, encryptkeys)
	nonzero := 0
	for i := uint(0); i < nHashToGenerate; i++ {
		if hashtable.Keys[i] != 0 {
			nonzero++
			//fmt.Printf("hash %x from key %x\n", hashtable.Keys[i], hashtable.Values[i])
		}
	}
	doneDict := time.Now()
	fmt.Printf("Fixed up dictionary in %v\n", doneDict.Sub(doneHashing))

	output := make(chan [2]uint, nThreads)
	//op := len(hashtable)
	for i := uint(0); i < nThreads*increment; i += increment {
		start := i | 0x0101010101010101
		go decryptWithAllKeys(start, nCheckPerThread, hashtable, output)

	}
	final := make([][2]uint, nThreads)
	for i := uint(0); i < nThreads; i++ {
		final[i] = <-output
	}
	fmt.Println(final)
}

func validate(encrypt, decrypt uint) []byte {

	key1 := intToString(encrypt)
	key2 := intToString(decrypt)

	fmt.Printf("key1 is %x, key2 is %x \n", getInt2(key1), getInt2(key2))

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

	validate(0x1c101010175cfb0, 0x161010101010104)

	//meetInTheMiddle()

}
