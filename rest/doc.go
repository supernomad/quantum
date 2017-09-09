// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

/*
Package rest contains the structs and logic to handle exposing internal metrics via a simple REST api for consumption by a myriad of different collection mechanisms.

The rest api is exposed by default at 'http://127.0.0.1:1099/metrics', but the ip, port, and uri are configurable at run time.

The statistics structure that is exposed is the following with both the links and queues objects being variable based on usage:
	{
	  "TxMetrics": {
	    "droppedPackets": 0,
	    "packets": 2,
	    "droppedBytes": 0,
	    "bytes": 40,
	    "links": {
	      "10.99.0.1": {
	        "droppedPackets": 0,
	        "packets": 2,
	        "droppedBytes": 0,
	        "bytes": 40
	      }
	    },
	    "queues": [
	      {
	        "droppedPackets": 0,
	        "packets": 2,
	        "droppedBytes": 0,
	        "bytes": 40
	      }
	    ]
	  },
	  "RxMetrics": {
	    "droppedPackets": 1,
	    "packets": 1,
	    "droppedBytes": 20,
	    "bytes": 20,
	    "links": {
	      "10.99.0.1": {
	        "droppedPackets": 0,
	        "packets": 1,
	        "droppedBytes": 0,
	        "bytes": 20
	      }
	    },
	    "queues": [
	      {
	        "droppedPackets": 1,
	        "packets": 1,
	        "droppedBytes": 20,
	        "bytes": 20
	      }
	    ]
	  }
	}
*/
package rest
