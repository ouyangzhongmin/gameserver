package object

import (
	"testing"
	"time"
)

func TestMonsterIdGen(t *testing.T) {

	for oid := 1; oid < 100; oid += 5 {
		//注意这里的oid 不能超过10万，否则offsetId会被截取为0
		offsetId := (oid * 10000) & 0xFFFFFFFF
		offsetTs := int64(time.Now().Unix() % 1000000)
		var id int64 = int64(offsetId) + offsetTs + int64(1)
		t.Log("mmid:", oid, offsetId, offsetTs, id)
	}

	for oid := 10000; oid < 100000; oid += 1000 {
		offsetId := (oid * 10000) & 0xFFFFFFFF
		offsetTs := int64(time.Now().Unix() % 1000000)
		var id int64 = int64(offsetId) + offsetTs + int64(1)
		t.Log("nnid:", oid, offsetId, offsetTs, id)
	}
}
