/*
    network server
    and in fact cache the network results into disk for fast recover
    and check the some validity: types check
    */
package main

import (
    pb "../protobuf/go"
    "strconv"
    //"../hash"
    "errors"
    "golang.org/x/net/context"
    "fmt"
    "google.golang.org/grpc"
	//"log"
	"sync"
	"time"
)

func checkUserID(UserID string)bool{
    return len(UserID)==8
}

/*func checkHash(blockHash string)bool{  //check first 5 chars?
    if len(blockHash) != 64{
        return false
    }
    return true
}*/

func checkTransaction(t *pb.Transaction)bool{
    if t.Value <= 0 || t.MiningFee <= 0{
        return false
    }
    if t.MiningFee >= t.Value{
        return false
    }
    if !checkUserID(t.FromID) || !checkUserID(t.ToID){
        return false
    }
	if t.FromID == t.ToID{
		return false
	}
    //t.FromID == t.ToID
    //T.Value > t.MiningFee
    //t.MiningFee > 0
    return true
}

func checkBlock(block *Block)bool{
    if len(block.Transactions) == 0{
        return false
    }
    if len(block.PrevHash)!=64{
        return false
    }
    if !checkUserID(block.MinerID){
        return false
    }
    if block.BlockID < 1{
        return false
    }
    if len(block.Nonce) != 8{
        return false
    }
    _, e := strconv.Atoi(block.Nonce)
    if e!=nil{
        return false
    }
    for _, t := range block.Transactions{
        if !checkTransaction(t){
            return false
        }
    }
    return true
}

func checkHashBlock(blockhash string, block *Block)bool{
    return checkBlock(block) && GetHashString(block.MarshalToString()) == blockhash
}

type Service struct{
    //To receive other's quest
    GetRequest chan string
    GetResponse chan int
    //TRANSFER //handeled by transfer
    VerifyRequest chan string
    VerifyResponse chan *Block
    //PushTransaction 
    GetHeightRequest chan bool
    GetHeightResponse chan *Block

    GetBlockRequest chan string
    GetBlockReponse chan *Block

    PushBlockRequest chan *Block
    PushBlockResponse chan bool

    Hello chan bool
	
	transfer *TransferManager
}

func NewService() *Service{
    s := &Service{}
    s.GetRequest = make(chan string)
    s.GetResponse = make(chan int)
    s.VerifyRequest = make(chan string)
    s.VerifyResponse = make(chan *Block)

    s.GetHeightRequest = make(chan bool)
    s.GetHeightResponse = make(chan *Block)

    s.GetBlockRequest = make(chan string)
    s.GetBlockReponse = make(chan *Block)

    s.PushBlockRequest = make(chan *Block)
    s.PushBlockResponse = make(chan bool)

    s.Hello = make(chan bool, 1)
    return s
}

func (s *Service) Get(q *pb.GetRequest) (*pb.GetResponse, error) {
    if !checkUserID(q.UserID){
        return &pb.GetResponse{Value: -1}, nil
    }
    s.GetRequest <- q.UserID
    return &pb.GetResponse{Value: int32(<-s.GetResponse)}, nil
}

/*
func (s *Service) Transfer(in *pb.Transaction) (*pb.BooleanResponse, error) {
    return &pb.BooleanResponse{Success: true}, nil
}
*/

func (s *Service) Transfer(in *pb.Transaction) (*pb.BooleanResponse, error) {
	if !checkTransaction(in){
		return &pb.BooleanResponse{Success: false}, nil
	}
	ok := s.transfer.AddPending(in)
	if ok{
		PushTransaction_client(in)
		return &pb.BooleanResponse{Success: true}, nil
	}
	return &pb.BooleanResponse{Success: false}, nil  //todo
}

func (s *Service) Verify(in *pb.Transaction) (*pb.VerifyResponse, error) {
    //We don't need to check other things
    if !checkTransaction(in){
        return &pb.VerifyResponse{Result: pb.VerifyResponse_FAILED, BlockHash:"?"}, nil
    }
    s.VerifyRequest <- in.UUID
    ok := <- s.VerifyResponse
    return &pb.VerifyResponse{Result: pb.VerifyResponse_FAILED, BlockHash:ok.GetHash()}, nil
	//todo SUCCEEDED PENDING
}

var Conn []*grpc.ClientConn
var ConnClient []pb.BlockChainMinerClient
var ConnStatus []int
var connLock sync.RWMutex
func CheckServer(){
	//check server
	for ;;{
		//fmt.Println("Checking Server")
		for i:=1; i<=int(Dat["nservers"].(float64)); i++{
			if i==IDstrInt{
				continue
			}
			connLock.RLock()
			if ConnStatus[i]==1{
				Conn[i].Close()
			}
			connLock.RUnlock()
			dat := Dat[strconv.Itoa(i)].(map[string]interface{})
			address, _ := fmt.Sprintf("%s:%s", dat["ip"], dat["port"]), fmt.Sprintf("%s",dat["dataDir"])
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			connLock.Lock()
			if err != nil {
				//log.Fatalf("Cannot connect to server: %v", err)
				ConnStatus[i] = 0
				Conn[i] = conn
			} else {
				ConnStatus[i] = 1
				Conn[i] = conn
				ConnClient[i] = pb.NewBlockChainMinerClient(Conn[i])
			}
			connLock.Unlock()
		}
		//time.Sleep(2 * time.Second)
		//fmt.Println("Checking Server Sleep")
		time.Sleep(10 * time.Second)
	}
}

/*func PushTransaction_client(in *pb.Transaction){
	//client
	for i:=1; i<=int(int(Dat["nservers"].(float64))); i++{
		if i==IDstrInt{
			continue
		}
		dat := Dat[strconv.Itoa(i)].(map[string]interface{})
		address, _ := fmt.Sprintf("%s:%s", dat["ip"], dat["port"]), fmt.Sprintf("%s",dat["dataDir"])
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Cannot connect to server: %v", err)
		}
		c := pb.NewBlockChainMinerClient(conn)
		r, err := c.PushTransaction(context.Background(), in)
		conn.Close()
	}
    return
}*/

func PushTransaction_client(in *pb.Transaction){
	for i:=1; i<=int(Dat["nservers"].(float64)); i++{
		connLock.RLock()
		s := ConnStatus[i]
		c := ConnClient[i]
		connLock.RUnlock()
		if s==1{
			c.PushTransaction(context.Background(), in)
		}
	}
	return
}

func (s *Service) PushTransaction(in *pb.Transaction) (*pb.Null, error) {
	if !checkTransaction(in){
		return &pb.Null{}, nil
	}
	ok := s.transfer.AddPending(in)
	if ok{
		PushTransaction_client(in)
	}
	return &pb.Null{}, nil
}

func (s *Service) PushBlock(in *pb.JsonBlockString) (*pb.Null, error) {
    block := MakeNewBlock()
    e := block.Unmarshal(in.Json)
    if e!=nil{
        return &pb.Null{}, e
    }
    if !checkBlock(block){
        return &pb.Null{}, errors.New("Invalid Json of PushBlock")
    }
    s.PushBlockRequest <- block
    <- s.PushBlockResponse
    return &pb.Null{}, nil
}

func (s *Service) GetHeight(in *pb.Null) (*pb.GetHeightResponse, error) {
    //return &pb.GetHeightResponse{Height: 1, LeafHash: "?"}, nil
	s.GetHeightRequest <- true
	block := <- s.GetHeightResponse
	height := block.BlockID
    return &pb.GetHeightResponse{Height: height, LeafHash: block.GetHash()}, nil
	
}

func (s *Service) GetBlock(in *pb.GetBlockRequest) (*pb.JsonBlockString, error) {
    if !CheckHash(in.BlockHash){
        return nil, errors.New("not a hash")
    }
    s.GetBlockRequest <- in.BlockHash
    block := <- s.GetBlockReponse
    if block == nil{
        return nil, nil
    } else{
        return &pb.JsonBlockString{Json:block.MarshalToString()}, nil
    }
}

type Server interface{
    GetHeight()(int, *Block, bool)
    GetBlock(hash string)(*Block, bool)
	PushBlock(block *Block, success chan bool)
    TRANSFER()*Transaction

    GetBlocksByBalance(*DatabaseEngine, chan *Block, chan int)
}

type RealServer struct{
    //rpc *grpc.Server
}

func (s *RealServer)GetBlock(hash string)(*Block, bool){
	for i:=1; i<=int(Dat["nservers"].(float64)); i++{
		connLock.RLock()
		status := ConnStatus[i]
		c := ConnClient[i]
		connLock.RUnlock()
		if status==1{
			res, e := c.GetBlock(context.Background(), &pb.GetBlockRequest{BlockHash:hash}) 
			if e==nil{
				block := MakeNewBlock()
				e := block.Unmarshal(res.Json)
				if e==nil{
					if checkHashBlock(hash, block){
						return block, true
					}
				}
				continue
				//fmt.Println("receive a strange hash value from someone else")
				//return nil, false
			}
		}
	}
	return nil, false
}

func (s *RealServer)GetHeight()(int, *Block, bool){
	for i:=1; i<=int(Dat["nservers"].(float64)); i++{
		connLock.RLock()
		status := ConnStatus[i]
		c := ConnClient[i]
		connLock.RUnlock()
		if status==1{
			res, e := c.GetHeight(context.Background(), &pb.Null{})
			if e!=nil {
				continue
			}
			block, ok := s.GetBlock(res.LeafHash)
			if ok{
				return int(block.BlockID), block, true
			}
		}
	}
    return -1, nil, false
}

func (s *RealServer)PushBlock(block *Block, success chan bool){
	json := block.MarshalToString()
    hash := block.GetHash()
    go WriteJson(hash, json)
	//for ;;{
	for i:=1; i<=int(Dat["nservers"].(float64)); i++{
		connLock.RLock()
		status := ConnStatus[i]
		c := ConnClient[i]
		connLock.RUnlock()
		if status==1{
			_, err := c.PushBlock(context.Background(), &pb.JsonBlockString{Json:json})
			if err!=nil{
				success<-true
			}
		}
	}
	//}
}

func (s *RealServer)TRANSFER()*Transaction{
    return &Transaction{}
}

func (s *RealServer)GetBlocksByBalance(*DatabaseEngine, chan *Block, chan int){
}
