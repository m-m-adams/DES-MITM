

def generate_keys(init):
    mask = 0x0101010101010101
    init = init |mask
    value = init
    while value < init+n//8:
        value = (value | mask) +1
        yield value

n = 1<<30
i=0
for start in range(0, n, n//8):
    a = generate_keys(start)
    print('round', i)
    i+=1
    print('starting at', hex(start | 0x0101010101010101))
    for key in a:
        last = hex(key)
    print ('ended at...', last)
