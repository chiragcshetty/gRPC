package main

import (
	"context"
	"flag"
	"log"
	"time"
	"fmt"
	"os"
	"bufio"
	//"strconv"
	//"math"
	"io/ioutil"
	//"strings"
	"io"
	"math/rand"
	"os/exec"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// 'kvexample' is main module name (i.e was initialized as 'go mod init kvexample')
	// Compiled protobuff are stored at '/kvstore/protocompiled/kvstore'
	kvs "kvexample/kvstore/protocompiled/kvstore_file"
)

//############################################## Globals ###################################################################################
var (
	storeServer kvs.StoreClient   // One client per server replica
	clientid           = flag.Int("id", 0, "client id")
	clientSignature float32            // eg: if clientid = 143, clientSignature is 0.143. 
									   // To have non-overlapping timestamps among clients
	// serverAddrInfo: file that contains replica info as "ip:port" - one per line
	serverAddrInfo     = flag.String("addrfile", "replicainfo.txt" , 
									"File containing server address in the format of host:port")
	logType        = flag.String("log", "off", `Logging Options: off (default), file 
									(stored at logs/server-port.txt), stdout`)
	workload   = flag.String("workload", "dataset/operations-0.dat", "File providing workload opertions. Defualt: dataset/operations-0.dat")
)

	
func check(e error) {
    if e != nil {
        panic(e)
    }
}

func getFile(key string) (string, int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testResponse, err := storeServer.GetTest(ctx, &kvs.Record{})
	log.Println(testResponse.GetFileChunk())

	fileStreamResponse, err := storeServer.Get(ctx, &kvs.Record{
		Key: key,
	})
	if err != nil {
		log.Println("error downloading:", err)
		return "", 2
	}
	fn := fmt.Sprintf("dataset_files/rcvd_files/file-%s.dat",key)
	f, err := os.Create(fn)
    check(err)
	defer f.Close()
	w := bufio.NewWriter(f)

	for {
		chunkResponse, err := fileStreamResponse.Recv()
		if err == io.EOF {
			log.Println("received all chunks")
			break
		}
		if err != nil {
			log.Println("err receiving chunk:", err)
			break
		}
		chunkLen, err := w.WriteString(chunkResponse.GetFileChunk())
    	check(err)
		log.Printf("wrote %d bytes\n", chunkLen)
	}
	w.Flush()

	return fn, 0
}

//###################################################### MAIN ###############################################################
func main() {

	flag.Parse()

	log.Printf("Starting. Logging Type %s, client ID %d ", *logType, *clientid ) 
	//******************************************** SETUP LOGGING **************************************************
	switch *logType {
	case "file":
			fmt.Sprintf("Logging type: %s, find logs at logs/client-%d.txt \n", *logType, *clientid) 
			file, err := os.OpenFile( fmt.Sprintf("./logs/client-%d.txt", *clientid), os.O_CREATE|os.O_WRONLY, 0666 )
			if err != nil { log.Printf("Failed to open log: ", err) }
			log.SetOutput(file); 
	case "stdout":
			log.SetOutput(os.Stdout)
	default: //off
			// source: https://stackoverflow.com/questions/10571182/how-to-disable-a-log-logger
			log.Printf("Logging switched off \n")
			log.SetFlags(0)
			log.SetOutput(ioutil.Discard) 
	}

	//************************************* SETUP: one client per replica *************************************************
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//--------------------------- Read replicas' info from serverAddrInfo file --------------------------
	// source: https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go
	conn, err := grpc.Dial("localhost:8008", opts...)
	if err != nil {
		log.Fatal("client could connect to grpc service:", err)
	}
	storeServer = kvs.NewStoreClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for i := 0; i < 12; i++ {
		_,_= storeServer.Set(ctx, &kvs.Record{
			Key: "KEY1006",
		})
	}

	cache := make(map[string]string)
	for i := 0; i < 50; i++  {
		j := rand.Intn(10)

		start := time.Now()
		key := fmt.Sprintf("%s%d", "KEY", 1000+j)
		fn, ok := cache[key]
		// If the key exists
		if ok {
			log.Println("Found in Cache")
		} else {
			log.Println("NOT Found in Cache")
			fn,_= getFile(key)
			if fn!=""{cache[key] = fn}
			
		}
		log.Println(fn)


 		_, err := exec.Command("cat", fn).Output()
		//log.Printf("The date is %s\n", outp)
		if err != nil {
			fmt.Printf("error %s \n", err)
		}
 

		duration := time.Since(start)
		log.Println(duration)
	}

}

