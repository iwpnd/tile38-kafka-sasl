// This script test the heap growth on a Tile38 server that has a bunch of
// SET/POINT operations all have random TTLs.
//
// The console will print server stats in real-time, and you should see the
// heap stay pretty stable over time.
//
// This script also executes an AOFSHRINK every 5 seconds or so.
//

package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gomodule/redigo/redis"
)

const minTTL = 30.0        // Min TTL for each SET operation in seconds
const maxTTL = 30.0        // Max TTL for each SET operation in seconds
const tile38Addr = ":9851" // Address of Tile38 Server

const nclients = 1 // 50
const plsize = 1
const nhooks = 100

func main() {
	rand.Seed(time.Now().UnixNano())

	var totalSets uint64
	var totalShrinks uint64
	// var totalNotifs uint64
	var totalClientErrs uint64

	// Add some Kafka hooks
	c := must(redis.Dial("tcp", tile38Addr)).(redis.Conn)
	defer c.Close()
	for i := 0; i < nhooks; i++ {
		x := rand.Float64()*360 - 180
		y := rand.Float64()*180 - 90
		if true {
			must(c.Do("SETHOOK", fmt.Sprintf("__tmphook__:%d", i),
				"kafka://broker:9091/test?auth=sasl&ssl=false&sha512=true",
				"INTERSECTS", "__tmpkey__",
				"DETECT", "enter,exit",
				"FENCE",
				"BOUNDS", x, y, x+5, y+5))
		} else {
			must(c.Do("SETCHAN", fmt.Sprintf("__tmpchan__:%d", i),
				"INTERSECTS", "__tmpkey__",
				"DETECT", "enter,exit",
				"BOUNDS", x, y, x+5, y+5))
		}

	}

	go func() {
		conn := must(redis.Dial("tcp", tile38Addr)).(redis.Conn)
		for {
			must(conn.Do("AOFSHRINK"))
			atomic.AddUint64(&totalShrinks, 1)
			time.Sleep(time.Second * 5)
		}
	}()

	for i := 0; i < nclients; i++ {
		go func() {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for {
				func() {
					c, err := redis.Dial("tcp", tile38Addr)
					if err != nil {
						atomic.AddUint64(&totalClientErrs, 1)
						return
					}
					defer c.Close()
					for {
						for j := 0; j < plsize; j++ {
							lat := rng.Float64()*180 - 90
							lon := rng.Float64()*360 - 180
							ttl := rng.Float64()*(maxTTL-minTTL) + (minTTL)
							id := randID(rng)
							must(nil, c.Send("SET", "__tmpkey__",
								id, "EX", ttl,
								"POINT", lat, lon))
							atomic.AddUint64(&totalSets, 1)
						}
						must(nil, c.Flush())
						for j := 0; j < plsize; j++ {
							must(c.Receive())
						}
					}
				}()
			}
		}()
	}
	go func() {
		var highHeap int
		conn := must(redis.Dial("tcp", tile38Addr)).(redis.Conn)
		lastMark := time.Now()
		for {
			vals := must(redis.StringMap(conn.Do("SERVER"))).(map[string]string)
			heap, _ := strconv.Atoi(vals["heap_size"])
			if heap > highHeap {
				highHeap = heap
			}
			nerrs := atomic.LoadUint64(&totalClientErrs)
			nobjs, _ := strconv.Atoi(vals["num_objects"])

			vals = must(redis.StringMap(conn.Do("SERVER", "EXT"))).(map[string]string)
			nhookgroups, _ := strconv.Atoi(vals["tile38_num_hook_groups"])
			nobjgroups, _ := strconv.Atoi(vals["tile38_num_object_groups"])

			fmt.Printf("\rheap: %.3f GB, points: %-10d groups: [%d:%d] ",
				float64(heap)/1024/1024/1024, nobjs,
				nhookgroups, nobjgroups,
			)
			if time.Since(lastMark) > time.Second*5 {
				fmt.Printf("-- (high heap: %.3f GB, client errs: %d)\n",
					float64(highHeap)/1024/1024/1024, nerrs)
				lastMark = time.Now()
			}
			time.Sleep(time.Second / 10)
		}
	}()
	select {}
}

func randID(rng *rand.Rand) string {
	var id [16]byte
	rng.Read(id[:])
	return hex.EncodeToString(id[:])
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}
