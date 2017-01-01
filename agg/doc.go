// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

/*
Package agg contains the structs and logic to handle aggregating transmission and reception statistics for quantum. These statistics are then exposed via a simple REST api for consumption by a myriad of different collection points.

All of the metrics collected by quantum are represented by monotonically incrementing counters, and will only reset in the event of an application restart or if the counter increments beyond the size of a uint64.

The rest api is exposed by default at 'http://127.0.0.1:1099/stats'.

The statistics structure that is exposed looks similar to the following:
	{
	  "TxStats": {
	    "droppedPackets": 0,
	    "packets": 2,
	    "droppedBytes": 0,
	    "bytes": 0,
	    "links": {
	      "10.99.0.1": {
	        "droppedPackets": 0,
	        "packets": 2,
	        "droppedBytes": 0,
	        "bytes": 0
	      }
	    },
	    "queues": [
	      {
	        "droppedPackets": 0,
	        "packets": 2,
	        "droppedBytes": 0,
	        "bytes": 0
	      }
	    ]
	  },
	  "RxStats": {
	    "droppedPackets": 1,
	    "packets": 0,
	    "droppedBytes": 0,
	    "bytes": 0,
	    "queues": [
	      {
	        "droppedPackets": 1,
	        "packets": 0,
	        "droppedBytes": 0,
	        "bytes": 0
	      }
	    ]
	  }
	}
*/
package agg
