# DES-MITM

Meet in the middle DES based hash brute force attacker for "weakhash" from Northsec 2020

has a hard coded plaintext as a base for the hash function, and a known ciphertext in the "decrypt with all keys" function

TODO: change this to accept a plaintext/ciphertext pair as commandline inputs for future use

Runs an MITM attack - first encrypts the plaintext by 2^30 possible keys and stores all hash:key pairs in a map
next decrypts the known ciphertext by all possible keys until it finds a match
The program is multithreaded and will run on 8 cores if available

NOTE - DO NOT RUN THIS IF YOU DON'T KNOW WHAT IT DOES! There are no checks on available processors or memory.
It will eat 32 gb of ram and 100% of 8 cores for 6ish hours. Be aware of this as you plan your day.
