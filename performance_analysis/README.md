
# Performance analysis

This directory contains tools for analyzing tapper performance.

# Periodic tapper logs

In tapper logs there are some periodic lines that shows its internal state and consumed resources.

Internal state example (formatted and commented):
```
stats - {
	"processedBytes":468940592, // how many bytes we read from pcap
	"packetsCount":174883, // how many packets we read from pcap
	"tcpPacketsCount":174883, // how many tcp packets we read from pcap
	"reassembledTcpPayloadsCount":66893, // how many chunks sent to tcp stream
	"matchedPairs":24821, // how many request response pairs found
	"droppedTcpStreams":2 // how many tcp streams remained stale and dropped
}
```

Consumed resources example (formatted and commented):
```
mem: 24441240, // golang heap size
goroutines: 29, // how many goroutines
cpu: 91.208791, // how much cpu the tapper process consume (in percentage per core)
cores: 16,  // how many cores there are on the machine
rss: 87052288 // how many bytes held by the tapper process
```

# Plot tapper logs

In order to plot a tapper log or many logs into a graph, use the `plot_from_tapper_logs.py` util.

It gets a list of tapper logs as a parameter, and output an image with a nice graph.

The log file names should be named in this format `XX_DESCRIPTION.log` when XX is the number between determining the color of the output graph and description is the name of the series. It allows for easy comparison between various modes.

Example run:
```
cd $MIZU_HOME/performance_analysis
virtualenv venv
source venv/bin/activate
pip install -r requirements.txt
python plot_from_tapper_logs.py 00_tapper.log
```

# Tapper Modes

Every packet seen by the tapper is processed in a pipeline that contains various stages. 
* Pcap - Read the packet from libpcap
* Assembler - Assemble the packet into a TcpStream
* TcpStream - Hold stream information and TcpReaders
* Dissectors - Read from TcpReader and recognize the packet content and protocol.
* Emit - Marshal the request response pair into a Json
* Send - Send the Json to Api Server

Tapper can be run with various debug modes:
* No Pcap - Start the tapper process, but don't read from any packets from pcap
* No Assembler - Read packets from pcap, but don't assemble them
* No TcpStream - Assemble the packets, but don't create TcpStream for them
* No Dissectors - Create a TcpStream for the packets, but don't dissect their content
* No Emit - Dissect the TcpStream, but don't emit the matched request response pair 
* No Send - Emit the request response pair, but don't send them to the Api Server.
* Regular mode

![Tapper Modes](https://github.com/up9inc/mizu/blob/debug/profile-tapper-benchmark/performance_analysis/tapper-modes.png)

# Run benchmark with various tapper modes

## Prerequisite

In order to run the benchmark you probably want:
1. An up and running Api Server
2. An up and running Basenine
3. An up and running UI (optional)
4. An up and running test server, like nginx, that can return a known payload at a known endpoint.
5. Set MIZU_HOME environment variable to points to mizu directory
6. Install the `hey` tool

## Running the benchmark

In order to run a benchmark use the `run_tapper_benchmark.sh` script.

Example run:
```
cd $MIZU_HOME/performance_analysis
source venv/bin/activate # Assuming you already run plot_from_tapper_logs.py 
./run_tapper_benchmark.sh
```

Running it without params use the default values, use the following environment variables for customization:
```
export=MIZU_BENCHMARK_OUTPUT_DIR=/path/to/dir # Set the output directory for tapper logs and graph
export=MIZU_BENCHMARK_CLIENT_PERIOD=1m # How long each test run
export=MIZU_BENCHMARK_URL=http://server:port/path # The URL to use for the benchmarking process (the test server endpoint)
export=MIZU_BENCHMARK_RUN_COUNT=3 # How many times each tapper mode should run
export=MIZU_BENCHMARK_QPS=250 # How many queries per second the each client should send to the test server
export=MIZU_BENCHMARK_CLIENTS_COUNT=5 # How many clients should run in parallel during the benchmark
```

# Example output graph

An example output graph from a 15 min run with 15K payload and 1000 QPS looks like

![Example Graph](https://github.com/up9inc/mizu/blob/debug/profile-tapper-benchmark/performance_analysis/example-graph.png)

