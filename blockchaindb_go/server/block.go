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
}

type Transaction struct{
    Type string
    FromID string
    ToID string
    Value int
    MiningFee int
    UUID string
}

type Block struct{
    BlockId int
    PrevHash string
    Transactions *[]Transaction
    MinerID string
    Nonce string

    Depth int //Depth of the block

    MyHash string
}

func (b *Block) CheckHash(data string)bool{
    bytes := hash.GetHashBytes(data)
    return bytes[0] == 0 && bytes[1] == 0 && bytes[2] == 0 && bytes[3] == 0 && bytes[4] == 0
} 

func (b *Block) MarshalToString()(string,error){
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
