
echo "Compiling kvstore protobuf"
cd kvstore
protoc --go_out=.   --go-grpc_out=.   *.proto

cd ..
echo "Building Client: client/kvclient.go "
go build client/kvclient.go
echo "Building Server: server/kvserver.go"
go build server/kvserver.go

echo "Success. Done!"
echo "Now, update server replica information in replicainfo.txt."
echo "Then run workload.sh"
