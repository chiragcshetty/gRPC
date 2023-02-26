import string
import random

f = open("./dataset.dat", "w+")

keysize = 24-4
valuesize = 10-4
datasetSize = 1000000

for i in range(datasetSize):
    key = str(i)
    key = ''.join(["0" for _ in range(keysize-len(key))]) + key
    value = ''.join(random.choices(string.ascii_letters, k=valuesize))
    f.write("KEY-{} VAL-{}\n".format(key,value))

f.close()


fo = open("./operations.dat", "w+")
numOperation = 2000
fractionRead = 0.5
fractionWrite = 1 -  fractionRead

for j in range(numOperation):
    
    key = str(random.randint(0,datasetSize))
    key = ''.join(["0" for _ in range(keysize-len(key))]) + key
    key = "KEY-"+key

    r_or_w = (random.random() < fractionRead) # 1 is read

    if r_or_w: #read
        fo.write("GET "+ key + "\n")
    else:
        value = ''.join(random.choices(string.ascii_letters, k=valuesize))
        value = "VAL-"+value
        fo.write("SET {} {} \n".format(key, value))

fo.close()