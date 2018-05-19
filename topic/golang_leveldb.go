package main

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	DB *leveldb.DB
)

type KV struct {
	Key   string
	Value string
}

func main() {
	var err error
	DB, err = leveldb.OpenFile("./db", nil)
	if err != nil {
		fmt.Printf("open file err %v", err)
		return
	}
	defer DB.Close()

	msg := &KV{
		Key:   "cowkeys001",
		Value: "valueabc",
	}

	if err := PutData(msg); err != nil {
		fmt.Printf("put data faield%v", err)
		return
	}

	GetData(msg)

}

func PutData(msg *KV) error {
	return DB.Put([]byte(msg.Key), []byte(msg.Value), nil)
}

func GetData(msg *KV) error {
	data, err := DB.Get([]byte(msg.Key), nil)
	if err != nil {
		fmt.Printf("=======error:%v", err)
		return nil
	}

	fmt.Printf("======get result:%v", string(data))
	return nil
}

