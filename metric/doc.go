// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

/*
Package metric contains the structs and logic to handle aggregating transmission and reception statistics for quantum.

All of the metrics collected by quantum are represented by monotonically incrementing uint64 counters, and will only reset in the event of an application restart or if the counters increment beyond the size of a uint64.

The currently collected metrics are as follows:
    - Packets
    - Dropped Packets
    - Bytes
    - Dropped Bytes

The metrics are split out based on the queue and the link that handled the transmission, as well as generally over all queues/links. Where a link represents the remote peer involved in the transmission, and a queue represents the internal packet queue.
*/
package metric
