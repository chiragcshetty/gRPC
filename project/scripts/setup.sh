cd ..

go mod tidy
echo "Compiling kvstore protobuf"
cd kvstore
protoc --go_out=.   --go-grpc_out=.   *.proto

cd ..
echo "Building Client: client/kvclient.go "
go build client/kvclient.go
echo "Building Server: server/kvserver.go"
go build server/kvserver_file.go

#echo "Generating dataset to initialize the kv store, See dataset/dataset.dat. And workload operations, See dataset/operations.dat"
#cd dataset
#python3 generate_dataset.py
#cd ..

rm -rf dataset_files/rcvd_files/*
rm -rf dataset_files/files/*

echo "Success. Done!"
echo " "
echo "Next steps:"

echo "1. Update server replica information in replicainfo.txt. Format replica_ip:port. One replica per line"
echo "2. Run server replica: Example: ' ./kvserver -port 8008 -log stdout '. '-log' can be off, stdout or file. If file is choosen, the log will be at logs/server-<port>.txt. The KV store is automatically initialized with data in dataset/dataset.dat"
echo "3. Run client: Example: ' ./kvclient -id 2 -log stdout -workload=./dataset/operations-2.dat'. Client's id must be an integer. Client runs the workload in dataset/operations.dat and prints out the time taken"

echo "To kill processes using needed ports: fuser -k 8008/tcp 8007/tcp 8006/tcp"


