package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"os"
	"io/ioutil"
	//"bufio"
	//"strings"
	"container/heap"
	"io"
	"os/exec"
	"time"
	"errors"
	//"math/rand"

	"google.golang.org/grpc"

	//'kvexample' is main module name (i.e was initialized as 'go mod init kvexample')
	// Compiled protobuff are stored at '/kvstore/protocompiled/kvstore'
	kvs "kvexample/kvstore/protocompiled/kvstore_file"
)

//########################################## Globals #######################################################################################
var (
	logType        = flag.String("log", "off", "Logging Options: off (default), file (stored at logs/server-port.txt), stdout ")
	port           = flag.String("port", "8009", "The server will listen on this port")
	//filesize       = flag.Int("filesize", 1000, "Size of file in bytes")
	//binsize        = flag.Int("size of garbage bin", 10, "number of files in garbage bin") // TODO: Change to size of bin in bytes
	hfilePath      = "dataset_files/hfiles/"
	hfileNo        = 0
	hfile          = fmt.Sprintf("%s%d.dat",hfilePath,hfileNo) 
	lastCompactHfile = 0
)
const(
	//binsize = 10
	//filesize = "1000"
	hfilesizelimit = 20  //number of keys in a hfile
	compactionlimit = 5  //number of files for compaction
	valsize   = "10"
)

//###########################################
//************************************************************************************

// An Item is something we manage in a priority queue.
type KVItem struct {
	value    string // The value of the item; arbitrary.
	priority int32    // The priority = Key
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

type PriorityQueue []*KVItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	//log.Println(i,j)
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*KVItem)
	item.index = n
	*pq = append(*pq, item)
}


func (pq *PriorityQueue) PrintPQ() {
	old := *pq
	n := len(old)
	fmt.Println("*********************")
	for i := 0; i < n; i++ {
		fmt.Println(old[i].priority, old[i].value)
	}
	fmt.Println("***********--------------**********")

}

func (pq *PriorityQueue) Pop() interface{} {
	//pq.PrintPQ()
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) fixHeap(item *KVItem) {
	heap.Fix(pq, item.index)
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) updatePriority(item *KVItem, priority int32) {
	item.priority = priority
	heap.Fix(pq, item.index)
}

func (pq *PriorityQueue) getValue(key int32) (string, error) {
	que := *pq
	n := len(que)
	for i := 0; i < n; i++ {
		if key == que[i].priority {
			return que[i].value, nil
		}
	}
	return "", errors.New("Key not found")
}


func check(e error) {
    if e != nil {
        panic(e)
    }
}




//************************************************************************************
//########################################## The Actual KV store at this server ############################################################
type storeServer struct {
	kvs.UnimplementedStoreServer
	memstore PriorityQueue
	recent_set map[int32]int //key --> hfile no mapping

	// All read, write require locking of entire KV store. BAD!
	lock         sync.Mutex   // memstore lock
	rsLock       sync.Mutex   // recent set lock
}



//####################################### gRPC Server Service implemmentation ###############################################################
// The Services are specified in kvstore/kvstore_file.proto

///************-------------------- GET Request Implementation -----------------------------**********************

// Source: https://stackoverflow.com/questions/58566016/how-can-i-serve-files-while-using-grpc

func searchKey(key int32, filename string) (string, error){
	kv, err := exec.Command("bash", "findKey.sh", fmt.Sprint(key), filename ).Output()
	if err != nil {
		return "", err
	} else if string(kv)=="" {
		return "", errors.New("Key not found")
	}
	return string(kv), nil
}

func (s *storeServer) runCompaction () {
	for{
		if ((hfileNo-lastCompactHfile)> compactionlimit){
			s.rsLock.Lock()
			hfileNo = hfileNo + 1
			_, err := exec.Command("bash", "compact.sh", hfilePath, fmt.Sprint(hfileNo)).Output()
			check(err)
			s.recent_set = make(map[int32]int)
			lastCompactHfile = hfileNo
			s.rsLock.Unlock()
			log.Printf( fmt.Sprint("Compaction at ", hfileNo))
		}
		time.Sleep(30 * time.Second)
	}
}

func (s *storeServer) GetTest(ctx context.Context, record *kvs.Record) (*kvs.Response, error) {

	//fmt.Println("Request rcvd")
	return &kvs.Response{FileChunk:"TestSuccessful",}, nil
}


func (s *storeServer) Get(record *kvs.Record, responseStream kvs.Store_GetServer) error {
    
	// 1. Check if key is in memstore. If present resturn the value
	// 2. If not in memstore, check the recent_set. If present there, 
	//     check if timestamp from request and the hfile# match.
	//     		a. If they do, return "Cache valid" msg
	//			b. Else, return the new hfile
	// 3. If not present, the key hfile has been compacted out. 
	//    Server can either search and send the compacted file or if it 
	//    is too large then just search for the key and send the value

	//bufferSize := 64 *1024 //64KiB, tweak this as desired
	bufferSize := 5192 //64KiB, tweak this as desired

	// 1.--------------------------------------------------------------------
	s.lock.Lock()
	value, ok := s.memstore.getValue(record.GetKey())
	s.lock.Unlock()
	if ok == nil {   // key found in memstore
		resp := &kvs.Response{
				FileChunk: value,
				Timestamp: -1 ,
			}
		err := responseStream.Send(resp)
		check(err)
		/* if err != nil {
				log.Println("error while sending chunk:", err)
				return err
			}	 */
	
	} else {
		
		s.rsLock.Lock()
		keyHfileNo, ok :=  s.recent_set[record.GetKey()]
		s.rsLock.Unlock()

		// 2.--------------------------------------------------------------------
		if ok { // Present in recent_set
					// 2.a.******************************************************
				if keyHfileNo == int(record.GetTimeStamp()) {  
					 // And timestamp matches, So valid file
					resp := &kvs.Response{   // Cache valid
						FileChunk: "",
						Timestamp: -3,
					}
					err := responseStream.Send(resp)
					check(err)

				} else {  
					// 2.b.******************************************************  
					// timestamp not valid, so send new hfile
					filename := fmt.Sprintf("%s%d.dat",hfilePath,keyHfileNo) 
					file, err := os.Open(filename) 
					check(err)
					defer file.Close()

					buff := make([]byte, bufferSize)
					for {
						bytesRead, err := file.Read(buff)
						if err != nil {
							if err != io.EOF {
								fmt.Println(err)
							}
							break
						}
						resp := &kvs.Response{
							FileChunk: string(buff[:bytesRead]),
							Timestamp: int32(keyHfileNo),
						}
						err = responseStream.Send(resp)
						//log.Println(resp.GetFileChunk())
						if err != nil {
							log.Println("error while sending chunk:", err)
							return err
						}
						}
					}

				// 3.--------------------------------------------------------------------
		 		} else {
					// Not present in recent set. So search for KV yourself
					if lastCompactHfile>0 {
						filename := fmt.Sprintf("%s%d.dat",hfilePath,lastCompactHfile)
						val, err := searchKey(record.GetKey(), filename)
						if err != nil{
							resp := &kvs.Response{
								FileChunk: "",     // Key not found
								Timestamp: -2,
							}
							err = responseStream.Send(resp)
						} else {
							resp := &kvs.Response{
								FileChunk: val,
								Timestamp: int32(lastCompactHfile),     
							}
							err = responseStream.Send(resp)
							check(err)
						}

					} else {
						resp := &kvs.Response{
							FileChunk: "",     // Key not found
							Timestamp: -2,
						}
						err := responseStream.Send(resp)
						check(err)
					}
					
		 		}
	}
    return nil
}


///************-------------------- SET Request Implementation -----------------------------************************
// Request: SET (key, propoosed_value, proposed_ts). If key doesn't exist, add it. In any case, Ack to the sender.
func (s *storeServer) Set(ctx context.Context, record *kvs.Record) (*kvs.AckMsg, error) {
	//log.Printf("!~! Got a Set request for: %s\n", record.String())
 
	newkey   := record.GetKey()
	newvalue := record.GetValue()
	

	log.Printf(fmt.Sprint(newkey))

	kv := &KVItem {value: newvalue,
		priority: newkey,
		index: 0}

	s.lock.Lock()
	heap.Push(&(s.memstore), kv)
	log.Printf(fmt.Sprint(s.memstore.Len()))
	

	if s.memstore.Len() >= hfilesizelimit {
		
		s.rsLock.Lock()
		hfileNo = hfileNo + 1
		hfile     = fmt.Sprintf("%s%d.dat",hfilePath,hfileNo)
		log.Printf(hfile)
		kvitemj := int32(-1)      // To avoid duplicates: Fix the memstore to not have duplicates
		for {
			if s.memstore.Len() > 0 {
				kvitemi, fine := heap.Pop(&s.memstore).(*KVItem)
				//log.Printf( fmt.Sprint((*kvitemi).priority))
				if fine{
					if (kvitemj != (*kvitemi).priority) {
						_, err := exec.Command("bash", "add_kv.sh", hfile, fmt.Sprint((*kvitemi).priority) , (*kvitemi).value).Output()
						check(err)
						kvitemj = (*kvitemi).priority
						
						s.recent_set[kvitemj] = hfileNo
					}
				} else {break}
			} else {break}
		}
		//log.Printf("################################")
		s.rsLock.Unlock()
	}
	s.lock.Unlock()
	
	return  &kvs.AckMsg{}, nil
}

//TODO: Garbage collector

//#################################################### Utilities #############################################################################

// loadSome data into the KVstore to begin with
func (s *storeServer) loadKV() {
	hfileSize      := 0
	
	for i := 0; i < 200; i++ {
		//keynew := fmt.Sprintf("%s%d", "KEY", i)
		_, err := exec.Command("bash", "add_randomkv.sh", hfile, fmt.Sprint(i), valsize ).Output()
		if err != nil {
			fmt.Printf("error %s \n", err)
		}
		hfileSize = hfileSize + 1
		if hfileSize > hfilesizelimit {
			s.rsLock.Lock()
			hfileNo = hfileNo + 1
			s.rsLock.Unlock()
			hfileSize = 0
			hfile     = fmt.Sprintf("%s%d.dat",hfilePath,hfileNo)
		}

		//kv1 := kvs.Record{
		//	Key: keynew ,
		//}
		//log.Println(kv1.GetKey())
		//log.Println(fn)
		//s.store[kv1.GetKey()] = fn
	}	
	log.Printf("Setup Complete.\n\n")
}


func newServer() *storeServer {
	s := &storeServer{ memstore: make(PriorityQueue, 0), recent_set : make(map[int32]int) }
	//s.loadKV()
	go s.runCompaction()
	return s
}


//################################################ MAIN ##################################################################


func main() {
	flag.Parse()
	log.Printf("Starting. Logging Type %s, Port no %s ", *logType, *port ) 
	//******************************************** SETUP LOGGING **************************************************
	
	switch *logType {
	case "file":
			fmt.Sprintf("Logging type: %s, find logs at logs/server-%s.txt \n", *logType, *port) 
			file, err := os.OpenFile( ("./logs/server-" + *port + ".txt"), os.O_CREATE|os.O_WRONLY, 0666 )
			if err != nil { log.Printf("Failed to open log: %s ", err) }
			log.SetOutput(file); 
	case "stdout":
			log.SetOutput(os.Stdout)
	default: //off
			// source: https://stackoverflow.com/questions/10571182/how-to-disable-a-log-logger
			fmt.Sprintf("Logging switched off \n")
			log.SetFlags(0)
			log.SetOutput(ioutil.Discard) 
	}
	
	//***************************************** SETUP TCP SERVER********************************************************

	log.Printf("Replica Server Port No: %s \n\n", *port)
	lis, err := net.Listen("tcp", ":" + *port)
	if err != nil { log.Fatalf("Failed to listen: %v", err) }

	//***************************************** SETUP gRPC SERVER*******************************************************
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	kvs.RegisterStoreServer(grpcServer, newServer())

	//************************************* gRPC server now listening*****************************************************
	log.Printf("~~~~~~~~~ The server is now listening! ~~~~~~~~~~~~~~ \n\n")
	grpcServer.Serve(lis)

}