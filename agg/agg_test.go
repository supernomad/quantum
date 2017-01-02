// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package agg

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/Supernomad/quantum/common"
)

func Example() {
	resp, err := http.Get("http://127.0.0.1:1099/stats")
	if err != nil {
		fmt.Println("Error getting statistics:", err.Error())
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error parsing response body:", err.Error())
		panic(err)
	}

	fmt.Println(string(body))
	// Output: {"TxStats":{"droppedPackets":0,"packets":2,"droppedBytes":0,"bytes":40,"links":{"10.99.0.1":{"droppedPackets":0,"packets":2,"droppedBytes":0,"bytes":40}},"queues":[{"droppedPackets":0,"packets":2,"droppedBytes":0,"bytes":40}]},"RxStats":{"droppedPackets":1,"packets":1,"droppedBytes":20,"bytes":20,"links":{"10.99.0.1":{"droppedPackets":0,"packets":1,"droppedBytes":0,"bytes":20}},"queues":[{"droppedPackets":1,"packets":1,"droppedBytes":20,"bytes":20}]}}
}

func TestAgg(t *testing.T) {
	agg := New(
		common.NewLogger(common.NoopLogger),
		&common.Config{
			StatsRoute:   "/stats",
			StatsPort:    1099,
			StatsAddress: "127.0.0.1",
			NumWorkers:   1,
		})
	agg.Start()
	agg.Aggs <- &common.Stat{
		Direction: common.OutgoingStat,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}
	agg.Aggs <- &common.Stat{
		Direction: common.OutgoingStat,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}
	agg.Aggs <- &common.Stat{
		Direction: common.IncomingStat,
		Dropped:   true,
		Bytes:     20,
	}
	agg.Aggs <- &common.Stat{
		Direction: common.IncomingStat,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}

	time.Sleep(2 * time.Second)
	_, err := http.Get("http://127.0.0.1:1099/stats")
	if err != nil {
		t.Fatal(err)
	}
	agg.Stop()
}
