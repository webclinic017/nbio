package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/taskpool"
)

var (
	success      uint64 = 0
	failed       uint64 = 0
	totalSuccess uint64 = 0
	totalFailed  uint64 = 0
)

func main() {
	flag.Parse()

	clientExecutePool := taskpool.NewMixedPool(1024, 1, 1024)
	engine := nbhttp.NewEngineTLS(nbhttp.Config{
		NPoller: runtime.NumCPU(),
	}, nil, nil, &tls.Config{}, clientExecutePool.Go)
	engine.InitTLSBuffers()

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	go func() {
		ok := true
		cli := &nbhttp.Client{
			Engine: engine,
		}
		for ok {
			cli.Do(nil, func(res *http.Response, err error) {
				if err != nil {
					atomic.AddUint64(&failed, 1)
					fmt.Println("Do failed:", err)
					ok = false
					return
				} else {
					atomic.AddUint64(&success, 1)
					// fmt.Println(res.Proto, res.StatusCode, res.Status)
				}
			})
			time.Sleep(time.Second / 10)
		}
	}()

	ticker := time.NewTicker(time.Second)
	for i := 1; true; i++ {
		<-ticker.C
		nSuccess := atomic.SwapUint64(&success, 0)
		nFailed := atomic.SwapUint64(&failed, 0)
		totalSuccess += nSuccess
		totalFailed += nFailed
		fmt.Printf("running for %v seconds, success: %v, totalSuccess: %v, failed: %v, totalFailed: %v\n", i, nSuccess, totalSuccess, failed, totalFailed)
	}
}
