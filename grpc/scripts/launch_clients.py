import os
import sys
numClients = int(sys.argv[1])
firstClient = int(sys.argv[2])


os.chdir("..")
for i in range(firstClient,numClients+firstClient+1):
    print("Launching Client ", i)
    os.system("./kvclient -id {} -log off -workload=./dataset/operations-{}.dat &".format(i,i))
