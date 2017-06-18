package main

import "crypto/sha256"
import "fmt"

func GetHashString(String string) string {
	return fmt.Sprintf("%x", GetHashBytes(String))
}

func GetHashBytes(String string) [32]byte {
	return sha256.Sum256([]byte(String))
}

func GetHashBytes_(b []byte) [32]byte {
	return sha256.Sum256(b)
}

func CheckHash(Hash string) bool {
	for i:=0; i<HashHardness; i++{
        if Hash[i]!='0'{
            return false
        }
    }
	//Hash[0:5]=="00000"
	return len(Hash) == 64
}
