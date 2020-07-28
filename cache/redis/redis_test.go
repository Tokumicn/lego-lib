package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"lego-lib/cache"
)

func TestRedisCache(t *testing.T) {
	bm, err := cache.NewCache("redis", `{"conn": "127.0.0.1:6379"}`)
	if err != nil {
		t.Error("init err")
	}
	timeoutDuration := 10 * time.Second
	if err = bm.Put("legodemo", 1, timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("legodemo") {
		t.Error("check err")
	}

	time.Sleep(11 * time.Second)

	if bm.IsExist("legodemo") {
		t.Error("check err")
	}
	if err = bm.Put("legodemo", 1, timeoutDuration); err != nil {
		t.Error("set Error", err)
	}

	if v, _ := redis.Int(bm.Get("legodemo"), err); v != 1 {
		t.Error("get err")
	}

	if err = bm.Incr("legodemo"); err != nil {
		t.Error("Incr Error", err)
	}

	if v, _ := redis.Int(bm.Get("legodemo"), err); v != 2 {
		t.Error("get err")
	}

	if err = bm.Decr("legodemo"); err != nil {
		t.Error("Decr Error", err)
	}

	if v, _ := redis.Int(bm.Get("legodemo"), err); v != 1 {
		t.Error("get err")
	}
	bm.Delete("legodemo")
	if bm.IsExist("legodemo") {
		t.Error("delete err")
	}

	//test string
	if err = bm.Put("legodemo", "author", timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("legodemo") {
		t.Error("check err")
	}

	if v, _ := redis.String(bm.Get("legodemo"), err); v != "author" {
		t.Error("get err")
	}

	//test GetMulti
	if err = bm.Put("astaxie1", "author1", timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("astaxie1") {
		t.Error("check err")
	}

	vv := bm.GetMulti([]string{"legodemo", "astaxie1"})
	if len(vv) != 2 {
		t.Error("GetMulti ERROR")
	}
	if v, _ := redis.String(vv[0], nil); v != "author" {
		t.Error("GetMulti ERROR")
	}
	if v, _ := redis.String(vv[1], nil); v != "author1" {
		t.Error("GetMulti ERROR")
	}

	// test clear all
	if err = bm.ClearAll(); err != nil {
		t.Error("clear all err")
	}
}

func TestCache_Scan(t *testing.T) {
	timeoutDuration := 10 * time.Second
	// init
	bm, err := cache.NewCache("redis", `{"conn": "127.0.0.1:6379"}`)
	if err != nil {
		t.Error("init err")
	}
	// insert all
	for i := 0; i < 10000; i++ {
		if err = bm.Put(fmt.Sprintf("legodemo%d", i), fmt.Sprintf("author%d", i), timeoutDuration); err != nil {
			t.Error("set Error", err)
		}
	}
	// scan all for the first time
	keys, err := bm.(*Cache).Scan(DefaultKey + ":*")
	if err != nil {
		t.Error("scan Error", err)
	}
	if len(keys) != 10000 {
		t.Error("scan all err")
	}

	// clear all
	if err = bm.ClearAll(); err != nil {
		t.Error("clear all err")
	}

	// scan all for the second time
	keys, err = bm.(*Cache).Scan(DefaultKey + ":*")
	if err != nil {
		t.Error("scan Error", err)
	}
	if len(keys) != 0 {
		t.Error("scan all err")
	}
}
