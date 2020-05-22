package main

import (
	"crypto/des"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	//"github.com/alecthomas/mph"
	"runtime/debug"
)

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
	//this line has the real data from the problem
	//data := [2][]byte{[]byte("\xda\x99\xd1\xea\x64\x14\x4f\x3e"), []byte("\x59\xa3\x44\x2d\x8b\xab\xcf\x84")}

	//this is the provided test case
	//data := [1][]byte{[]byte("\xf3\x15\x06\x47\x12\x20\xcd\x8f")}

	//this is my test case with encode by 241 and decode by 123
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

func encryptWithAllKeys(start uint, nToGenerate uint, c chan map[uint]uint) {
	r := make(map[uint]uint, nToGenerate)
	plain := []byte("weakhash")
	mask := uint(0x01010101)
	counter := start

	for i := uint(0); i < nToGenerate; i++ {
		counter = (counter | mask) + 1
		key := intToString(counter)
		result := singlehash(plain, key)
		store := getInt2(result[:8])

		r[store] = counter
		i++
	}
	c <- r
	fmt.Printf("generated %d hashes from %x to %x \n", nToGenerate, start, counter)
}

func decryptWithAllKeys(start uint, nToGenerate uint, hashtable map[uint]uint, c chan [2]uint) {
	//data := []byte(0x59a3442d8babcf84)
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
			store := getInt2(dehashed[:8])

			if k, ok := hashtable[store]; ok {
				fmt.Printf("Key generated from int %x encrypt and int %x found for hash %x \n", k, counter, hash)
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

	nHashToGenerate := uint(1 << 29)
	nHashToCheck := uint(1 << 36)
	nThreads := uint(8)

	nGenPerThread := nHashToGenerate / nThreads
	nCheckPerThread := nHashToCheck / nThreads
	c := make(chan map[uint]uint, nThreads)
	fmt.Println("starting threads")

	increment32 := uint(1 << 29)
	increment64 := uint(1 << 36)
	for i := uint(0); i < nThreads*increment32; i += increment32 {
		start := i

		fmt.Printf("hash gen thread for %x to %x started\n", start, start+8*nGenPerThread)
		go encryptWithAllKeys(start, nGenPerThread, c)
	}

	hashtable := make(map[uint]uint, nHashToGenerate)
	for i := uint(0); i < nThreads; i++ {
		r := <-c
		for k, v := range r {
			hashtable[k] = v
			delete(r, k)
		}
		r = nil
	}

	fmt.Printf("generated all %d hashes\n", len(hashtable))

	output := make(chan [2]uint, nThreads)
	//op := len(hashtable)
	for i := uint(0); i < nThreads*increment64; i += increment64 {
		start := i

		fmt.Printf("hash check thread for %x to %x started\n", start, start+8*nCheckPerThread)
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
