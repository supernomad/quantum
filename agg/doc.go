// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

/*
Package agg contains the structs and logic to handle aggregating transmission and reception statistics for quantum. These statistics are then exposed via a simple REST api for consumption by a myriad of different collection mechanisms.

All of the metrics collected by quantum are represented by monotonically incrementing counters, and will only reset in the event of an application restart or if the counter increments beyond the size of a uint64.

The rest api is exposed by default at 'http://127.0.0.1:1099/stats', but the ip, port, and uri are configurable at run time. However the REST api does not currently support TLS encryption of the stats endpoint.

The statistics structure that is exposed is the following with both the links and queues objects being variable based on usage:
	{
	  "TxStats": {
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
	  "RxStats": {
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
package agg
