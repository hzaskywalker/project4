/*
    network server
    and in fact cache the network results into disk for fast recover
    and check the some validity: types check
    */
package main

import (
    pb "../protobuf/go"
    "strconv"
    "../hash"
    "errors"
    "golang.org/x/net/context"
    "fmt"
)

func checkUserID(UserID string)bool{
    return len(UserID)==8
}

func checkHash(blockHash string)bool{
    if len(blockHash) != 64{
        return false
    }
    return true
}

func checkTransaction(t *pb.Transaction)bool{
    if t.Value < 0 || t.MiningFee < 0{
        return false
    }
    if t.MiningFee > t.Value{
        return false
    }
    if !checkUserID(t.FromID) || !checkUserID(t.ToID){
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
    return checkBlock(block) && hash.GetHashString(block.MarshalToString()) == blockhash
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

func (s *Service) Verify(in *pb.Transaction) (*pb.VerifyResponse, error) {
    //We don't need to check other things
    if !checkTransaction(in){
        return &pb.VerifyResponse{Result: pb.VerifyResponse_FAILED, BlockHash:"?"}, nil
    }
    s.VerifyRequest <- in.UUID
    ok := <- s.VerifyResponse
    return &pb.VerifyResponse{Result: pb.VerifyResponse_FAILED, BlockHash:ok.GetHash()}, nil
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
    //need broad cast
    return &pb.Null{}, nil
}

func (s *Service) GetBlock(in *pb.GetBlockRequest) (*pb.JsonBlockString, error) {
    if !checkHash(in.BlockHash){
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

func (s *Service) GetHeight(in *pb.Null) (*pb.GetHeightResponse, error) {
    //return &pb.GetHeightResponse{Height: 1, LeafHash: "?"}, nil
	s.GetHeightRequest <- true
	block := s.GetHeightResponse
	height := block.BlockID
    return &pb.GetHeightResponse{Height: height, LeafHash: block.GetHash()}, nil
	
}

func (s *Service) Transfer(in *pb.Transaction) (*pb.BooleanResponse, error) {
	s.transfer.AddPending(in)
}

type Server interface{
    GetHeight()(int, *Block, bool)
    GetBlock(hash string)(*Block, bool)
    TRANSFER()*Transaction

    GetBlocksByBalance(*DatabaseEngine, chan *Block, chan int)
}

type RealServer struct{
    rpc *server
    ctx context.Context
}

func (s *RealServer)GetBlock(hash string)(*Block, bool){
    res, e := s.rpc.GetBlock(s.ctx, &pb.GetBlockRequest{BlockHash:hash}) 
    if e!=nil{
        return nil, false
    } else{
        block := MakeNewBlock()
        e := block.Unmarshal(res.Json)
        if e==nil{
            return block, true
        }
        fmt.Println("receive a strange hash value from someone else")
        return nil, false
    }
}

func (s *RealServer)GetHeight()(int, *Block, bool){
    res, e := s.rpc.GetHeight(s.ctx, &pb.Null{})
    if e!=nil {
        return -1, nil, false
    }
    block, ok := s.GetBlock(res.LeafHash)
    if ok{
        return int(block.BlockID), block, true
    }
    return -1, nil, false
}

func (s *RealServer)PushBlock(block *Block, success chan bool){
    json := block.MarshalToString()
    hash := block.GetHash()
    go WriteJson(hash, json)
    s.rpc.PushBlock(s.ctx, &pb.JsonBlockString{Json:json})
}

func (s *RealServer)TRANSFER()*Transaction{
    return &Transaction{}
}
