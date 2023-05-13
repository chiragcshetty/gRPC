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
	//"os/exec"
	//"errors"
	"os/exec"
	"sync"

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
	cache = make(map[int32]int)
	cacheLock sync.Mutex
	rcvdHfilePath      = "./dataset_files/rcvd_hfiles/"

)

type getResult struct {
	kv   string 
	returnCode int
}
	
func check(e error) {
    if e != nil {
        panic(e)
    }
}


func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
	   return false
	}
	return !info.IsDir()
 }


func getFromFile(randkey int32, filepath string) string{
	kvr, err := exec.Command("bash", "searchKeyClient.sh", fmt.Sprint(randkey), filepath).Output()
	if err != nil{ //not found
		log.Println("@@@@@@@@@@@@@@@@@")
		log.Println(fmt.Sprint(randkey))
		log.Println(filepath)
		log.Println(kvr)
		log.Println(err)
		log.Println("@@@@@@@@@@@@@@@@@")
		//return ""
	}
	check(err)
	return string(kvr)
}

// 1 means valid KV found
// 0 means key not found
// 2 means key found in cache. Wait for validity
// >2 wait for the other means

func getFromCache(randkey int32,  ch chan<- *getResult){
	cacheLock.Lock()
	cacheFileNo, ok := cache[randkey]
	cacheLock.Unlock()
	if ok{ //if key present in cache
			filepath := fmt.Sprintf("%s%d.dat",rcvdHfilePath,cacheFileNo)
			kvr := getFromFile(randkey, filepath)
			log.Println("Found in the cache!")
			ch <- &getResult{ kv:kvr , returnCode: 2 }

		} else { //if not in cache, search in all cache files
			ch <- &getResult{ kv:"", returnCode: 3 } //wait for server
/* 			files, err := ioutil.ReadDir(rcvdHfilePath)
			check(err)
			for _, file := range files {
				log.Println(fmt.Sprintf("%s%s.dat",rcvdHfilePath,file.Name()))
				kvr := getFromFile(randkey, fmt.Sprintf("%s%s.dat",rcvdHfilePath,file.Name()))
				if kvr!="" {
					ch <- &getResult{ kv:kvr , returnCode: -1*int(file.Name()[0]) } // TODO: update cache[randkey]
				}
			} */
		}		
}

func getFromServer(randkey int32, ch chan<- *getResult){
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ts := 0
	cacheLock.Lock()
	cache_ts, ok := cache[randkey]
	cacheLock.Unlock()
	if ok {ts=cache_ts}

	fileStreamResponse, err := storeServer.Get(ctx, &kvs.Record{
		Key: randkey,
		TimeStamp: int32(ts),
	})
	check(err)


	chunkResponse, err := fileStreamResponse.Recv()
	if err == io.EOF {
		log.Println("received all chunks")
	}
	if err != nil {
		log.Println("err receiving chunk:", err)
		return
	}

	//log.Printf(fmt.Sprint("TIMESTAMP:",int32(chunkResponse.GetTimestamp())))
	//log.Printf(fmt.Sprint("\n",string(chunkResponse.GetFileChunk())))

	replyTimeStamp := int(chunkResponse.GetTimestamp())
	replyKV := string(chunkResponse.GetFileChunk())

	if replyTimeStamp == -2 { //key not found
		log.Printf("Key not found")
		ch <- &getResult{ kv:"" , returnCode: 0 }

	} else if replyTimeStamp == -1 { //only value sent
		log.Printf("Key found in memstore")
		log.Printf(replyKV)
		ch <-&getResult{ kv: replyKV , returnCode: 1 }

	} else if replyTimeStamp == -3 { //valid cache  //wait for local cache read
		log.Printf("Valid cache")
		ch <- &getResult{ kv: "" , returnCode: 4 }

	} else { //new hfile sent
		//create a new hfile and update cache for that key

		fn := fmt.Sprintf("%s%d.dat",rcvdHfilePath,replyTimeStamp)

		if !fileExists(fn){
			log.Printf("New Hfile created")
			f, err := os.Create(fn)
			check(err)
			defer f.Close()
			w := bufio.NewWriter(f)

			chunkLen, err := w.WriteString(chunkResponse.GetFileChunk())
			check(err)
			log.Printf("wrote %d bytes\n", chunkLen)

			for{
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
		}
		log.Printf("Getting from the HFile")
		kvr := getFromFile(randkey, fn)
		ch <- &getResult{ kv: kvr , returnCode: 1 }
		cacheLock.Lock()
		cache[randkey] = replyTimeStamp
		cacheLock.Unlock()
			
		}
}


func getKV(randkey int32) string{
	getChannel := make(chan *getResult) 
	found :=""
	go getFromServer(randkey, getChannel)
	go getFromCache(randkey, getChannel)
	
	for{
		select {
			case r := <-getChannel:
				if r.returnCode==0 { return ""   // Key not found
				} else if r.returnCode==1 {return r.kv  // Key found
				} else if r.returnCode==2 { if found=="" { found=r.kv; log.Printf("waiting for valid confirmation ") } else {return r.kv} //Key in cache, but wait for server confirmation
				} else if r.returnCode==4 { if found=="" {found="validcache"; log.Printf("waiting for local cache out ")} else {return found} } // Sever confirms cache is valid

			default:
				time.Sleep(20 * time.Nanosecond)
		}
	}

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

	start := time.Now()
	for i := 0; i < 20000; i++ {
		randkey := int32(rand.Intn(300))
		log.Printf(fmt.Sprint("---------- KEY: ",randkey, " --------------"))
		getKV(randkey)
		log.Printf("------------------------************------------------------")
	}
	duration := time.Since(start)
	fmt.Println("Total time taken: ")
	fmt.Println(duration)

}

/* func getFile(key string) (string, int) {
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
} */

/* 	cache := make(map[string]string)   // Key to hfile
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
	} */