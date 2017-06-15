package main

/*
data structure for a block, need:

1. transformation between json and the structure
2. calculate the current hash value
3. solve for Nonce
*/

import {
    "../hash"
    pb "../protobuf/go"
    "github.com/golang/protobuf/jsonpb"
    "github.com/golang/protobuf/proto"
    "strings"
}

type Transaction struct{
    Type string
    FromID string
    ToID string
    Value int
    MiningFee int
    UUID string
    flag int //state of the transaction, sucess(0), pending(1), or not in longest (2)
}

type Block struct{
    BlockId int
    PrevHash string
    Transactions *[]Transaction
    MinerID string
    Nonce string

    Depth int //Depth of the block

    //value that should be decided when inserted into our databse
    MyHash string
    MyHeight int
}

func MakeNewBlock()*Block{
}

func (b *Block) GetHeight()int{
    return b.MyHeight
}

func (b *Block) SetHeight(height int){
    b.MyHeight = height
}


func (b *Block) GetHash() (string, error){
    //Maybe I need parallel this part
    if b.MyHash != ""{
        //store the hash value
        return b.MyHash
    }
    data, e := b.MarshalToString()
    if e== nil{
        b.MyHash = hash.GetHashString(data)
    }
    return b.MyHash, e
}

func CheckHashBytes(bytes []byte)bool{
    //Not sure
    return bytes[0] == 0 && bytes[1] == 0 && bytes[2] == 0 && bytes[3] == 0 && bytes[4] == 0
} 

func CheckHashString(a []string)bool{
    //Not sure
    return a[0] == "0" && a[1] == "0" && a[2] == "0" && a[3] == "0" && a[4] == "0" && a[5] == "0"
} 


func (b *Block) CheckHash() bool{
    a, e := b.GetHash()
    if e!=nil{
        return false
    }
    return CheckHashString(a)
}

func (b *Block) MarshalToString()string{
    //I don't think there would be error here
    block := new(pb.Block)
    block.BlockID = b.BlockID
    block.PrevHash = b.PrevHash
    block.Nonce = b.Nonce
    block.MinerID = b.MinerID
    for idx, i := range b.Transactions{
        block.Transactions[idx] = i
    }
    return jsonpb.MarshalToString(block), nil
}

func (b *Block) Unmarshal(data *string)error{
    block := new(pb.Block)
    err = jsonpb.UnmarshalString(data, block)

    b.Block = block.BlockID
    b.PrevHash = block.PrevHash
    b.Nonce = block.Nonce
    b.MinerID = block.MinerID

    //hard code here
    b.Transactions = make([]Transaction, 0, 50)
    for idx, i := range block.Transactions{
        b[idx] = i
    }
    return err, nil
}

func (b* Block) Solve(stop chan int, solved chan int){
    b.Nonce = "XXXXXXXX"
    data := b.MarshalToString()
    index := strings.index(data, b.Nonce)
    for i:=0;i<=99999999;++i{
        if (i&(0x11111))==0{
            //time out
            select {
                case res := <- stop:
                    if res{
                        return
                    }
                case <-time.After(time.Second * 0.0001):
                    fmt.Println("timeout 2")
            }
        }
        newNonce := fmt.Sprintf("%08x", i)
        data[index:index+8] = newNonce
        hashVal := hash.GetHashString(data)
        if CheckHashString(hashVal){
            b.MyHash = hashVal
            solved <- 1
            return
        }
    }
}