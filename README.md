# DES-MITM

Meet in the middle double DES cracker. It uses the built in map type to track hash/key pairs, which has a 16 byte overhead. Normally this would be fine, but it's very poor memory usage when trying to store billions of 8 byte values.  

The global variables control the amount of the hash that needs to match, the known plaintext/hash combinations, and the number of initial keys to generate. The defaults are set to match the weakhash problem from 2020 northsec. With the mitm.go in the main folder, it will use 16 * nKeys bytes of memory. With the map based version in the subfolder, it will use 32 * nKeys bytes of memory, but run slightly faster. There is a nearly linear inverse relationship between the time it takes to get a match and the amount of space you give the program.  (It's O(log(n)/n) time and O(n) space)


As configured it will eat 32 gb of ram and 100% of 8 cores for 20-30 minutes

DES drops the LSB of each byte, which effectively means that only 1/8 integers is a valid DES key. To get around this, the keys are generated with 
counter = counter | 0x0101010101010101 +1  
  
DES drops the LSB of each byte, so this loops through all possible DES keys without duplication. Basically every odd bite is converted to an even byte, and then 1 is added. There will be a single duplication every time a middle byte is even, and then it will correct in the next step. keygentest.py gives a visual demonstration of how it works.


