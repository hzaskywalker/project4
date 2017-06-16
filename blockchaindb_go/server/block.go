package main

/*
data structure for a block, need:

1. transformation between json and the structure
2. calculate the current hash value
3. solve for Nonce
*/

import (
    "../hash"
    pb "../protobuf/go"
    "github.com/golang/protobuf/jsonpb"
    "strings"
    "os"
    "fmt"
    "time"
)

var HashHardness int
var InitHash string

type Block struct{
    MyHash string
    pb.Block
}

func MakeNewBlock()*Block{
    return &Block{}
}

func (b *Block) GetHeight()int{
    //need to check ID
    return int(b.BlockID)
}


func (b *Block) GetHash()string{
    //Maybe I need parallel this part
    if b.MyHash != ""{
        //store the hash value
        return b.MyHash
    }
    data:= b.MarshalToString()
    b.MyHash = hash.GetHashString(data)
    return b.MyHash
}

func CheckHashBytes(bytes []byte)bool{
    for i:=0; i<HashHardness; i++{
        if bytes[i]!='0'{
            return false
        }
    }
    return true
    return bytes[0] == '0' && bytes[1] == '0' && bytes[2] == '0' && bytes[3] == '0' && bytes[4] == '0'
} 

func CheckHashString(a string)bool{
    //Not sure
    for i:=0; i<HashHardness; i++{
        if a[i]!='0'{
            return false
        }
    }
    return true
    return a[0] == '0' && a[1] == '0' && a[2] == '0' && a[3] == '0' && a[4] == '0' && a[5] == '0'
} 


func (b *Block) CheckHash() bool{
    a := b.GetHash()
    return CheckHashString(a)
}

func (b *Block) MarshalToString()string{
    //I don't think there would be error here
    block := new(pb.Block)
    block.BlockID = b.BlockID
    block.PrevHash = b.PrevHash
    block.Nonce = b.Nonce
    block.MinerID = b.MinerID
    block.Transactions = make([]*pb.Transaction, 0)
    for _, i := range b.Transactions{
        block.Transactions = append(block.Transactions, i)
    }
    t, e := (&jsonpb.Marshaler{EnumsAsInts: false}).MarshalToString(block)
    if e!=nil{
        fmt.Println("On block, marshall to string, error: ", e)
        os.Exit(1)
    }
    return t
}

func (b *Block)String()string {
    return b.MarshalToString()
}

func (b *Block) Unmarshal(data string){
    block := new(pb.Block)
    e := jsonpb.UnmarshalString(data, block)
    if e!=nil{
        fmt.Print(e)
        os.Exit(1)
    }

    b.BlockID = block.BlockID
    b.PrevHash = block.PrevHash
    b.Nonce = block.Nonce
    b.MinerID = block.MinerID

    //hard code here
    for _, i := range block.Transactions{
        b.Transactions = append(b.Transactions, i)
    }
    return
}

func (b* Block) Solve(stop chan int, solved chan int){
    b.Nonce = "XXXXXXXX"
    data := b.MarshalToString()
    index := strings.Index(data, b.Nonce)
    data_list := []byte(data)
    for i:=0;i<=99999999;i++{
        if (i&(0x11111))==0{
            //time out
            select {
                case res := <- stop:
                    if res==1 {
                        return
                    }
                case <-time.After(time.Nanosecond * 100):
                    {}
            }
        }
        newNonce := fmt.Sprintf("%08x", i)
        for j:=0;j<8;j++{
            data_list[index+j] = newNonce[j]
        }
        hashVal := hash.GetHashString(string(data_list))
        if CheckHashString(hashVal){
            b.Nonce = newNonce
            //fmt.Println(b.GetHash())
            b.MyHash = hashVal
            solved <- 1
            return
        }
    }
}

func (b *Block) SolveSync()string{
    stop := make(chan int)
    solved := make(chan int)
    go b.Solve(stop, solved)
    <- solved
    return b.MyHash
}