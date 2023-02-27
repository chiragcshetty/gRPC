import string
import random

########################################
numOperation = 20000
numClients   = 3
fractionRead = 0.5
fractionWrite = 1 -  fractionRead

keysize = 24-4
valuesize = 10-4 # '-4' for 'KEY-' or 'VAL-'
datasetSize = 1000000 #1M KV pairs

########################################

f = open("./dataset.dat", "w+")

print("Generating intial dataset.")
for i in range(datasetSize):
    key = str(i)
    key = ''.join(["0" for _ in range(keysize-len(key))]) + key
    value = ''.join(random.choices(string.ascii_letters, k=valuesize))
    f.write("KEY-{} VAL-{}\n".format(key,value))

f.close()


print("Generating workloads.")
for clientId in range (numClients+1):
    print("Generating workload for client:", clientId)
    fo = open("./operations-{}.dat".format(clientId), "w+")

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