#!/bin/bash

[ -z "$KUBESHARK_HOME" ] && { echo "KUBESHARK_HOME is missing"; exit 1; }
[ -z "$KUBESHARK_BENCHMARK_OUTPUT_DIR" ] && export KUBESHARK_BENCHMARK_OUTPUT_DIR="/tmp/kubeshark-benchmark-results-$(date +%d-%m-%H-%M)"
[ -z "$KUBESHARK_BENCHMARK_CLIENT_PERIOD" ] && export KUBESHARK_BENCHMARK_CLIENT_PERIOD="1m"
[ -z "$KUBESHARK_BENCHMARK_URL" ] && export KUBESHARK_BENCHMARK_URL="http://localhost:8081/data/b.1000.json"
[ -z "$KUBESHARK_BENCHMARK_RUN_COUNT" ] && export KUBESHARK_BENCHMARK_RUN_COUNT="3"
[ -z "$KUBESHARK_BENCHMARK_QPS" ] && export KUBESHARK_BENCHMARK_QPS="500"
[ -z "$KUBESHARK_BENCHMARK_CLIENTS_COUNT" ] && export KUBESHARK_BENCHMARK_CLIENTS_COUNT="5"

function log() {
	local message=$@
	printf "[%s] %s\n" "$(date "+%d-%m %H:%M:%S")" "$message"
}

function run_single_bench() {
	local mode_num=$1
	local mode_str=$2

	log "Starting ${mode_num}_${mode_str} (runs: $KUBESHARK_BENCHMARK_RUN_COUNT) (period: $KUBESHARK_BENCHMARK_CLIENT_PERIOD)"

	for ((i=0;i<"$KUBESHARK_BENCHMARK_RUN_COUNT";i++)); do
		log "  $i: Running tapper"
		rm -f tapper.log
		tapper_args=("--tap" "--api-server-address" "ws://localhost:8899/wsTapper" "-stats" "10" "-ignore-ports" "8899,9099")
		if [[ $(uname) == "Darwin" ]]
		then
			tapper_args+=("-i" "lo0" "-"decoder "Loopback")
		else
			tapper_args+=("-i" "lo")
		fi
		nohup ./agent/build/kubesharkagent ${tapper_args[@]} > tapper.log 2>&1 &

		log "  $i: Running client (hey)"
		hey -z $KUBESHARK_BENCHMARK_CLIENT_PERIOD -c $KUBESHARK_BENCHMARK_CLIENTS_COUNT -q $KUBESHARK_BENCHMARK_QPS $KUBESHARK_BENCHMARK_URL > /dev/null || return 1

		log "  $i: Killing tapper"
		kill -9 $(ps -ef | grep agent/build/kubesharkagent | grep tap | grep -v grep | awk '{ print $2 }') > /dev/null 2>&1

		local output_file=$KUBESHARK_BENCHMARK_OUTPUT_DIR/${mode_num}_${mode_str}_${i}.log
		log "  $i: Moving output to $output_file"
		mv tapper.log $output_file || return 1
	done
}

function generate_bench_graph() {
	cd performance_analysis/ || return 1
	source venv/bin/activate
	python plot_from_tapper_logs.py $KUBESHARK_BENCHMARK_OUTPUT_DIR/*.log || return 1
	mv graph.png $KUBESHARK_BENCHMARK_OUTPUT_DIR || return 1
}

mkdir -p $KUBESHARK_BENCHMARK_OUTPUT_DIR
rm -f $KUBESHARK_BENCHMARK_OUTPUT_DIR/*
log "Writing output to $KUBESHARK_BENCHMARK_OUTPUT_DIR"

cd $KUBESHARK_HOME || exit 1

export HOST_MODE=0
export SENSITIVE_DATA_FILTERING_OPTIONS='{}'
export KUBESHARK_DEBUG_DISABLE_PCAP=false
export KUBESHARK_DEBUG_DISABLE_TCP_REASSEMBLY=false
export KUBESHARK_DEBUG_DISABLE_TCP_STREAM=false
export KUBESHARK_DEBUG_DISABLE_NON_HTTP_EXTENSSION=false
export KUBESHARK_DEBUG_DISABLE_DISSECTORS=false
export KUBESHARK_DEBUG_DISABLE_EMITTING=false
export KUBESHARK_DEBUG_DISABLE_SENDING=false

export KUBESHARK_DEBUG_DISABLE_PCAP=true
run_single_bench "01" "no_pcap" || exit 1
export KUBESHARK_DEBUG_DISABLE_PCAP=false

export KUBESHARK_DEBUG_DISABLE_TCP_REASSEMBLY=true
run_single_bench "02" "no_assembler" || exit 1
export KUBESHARK_DEBUG_DISABLE_TCP_REASSEMBLY=false

export KUBESHARK_DEBUG_DISABLE_TCP_STREAM=true
run_single_bench "03" "no_tcp_stream" || exit 1
export KUBESHARK_DEBUG_DISABLE_TCP_STREAM=false

export KUBESHARK_DEBUG_DISABLE_NON_HTTP_EXTENSSION=true
run_single_bench "04" "only_http" || exit 1
export KUBESHARK_DEBUG_DISABLE_NON_HTTP_EXTENSSION=false

export KUBESHARK_DEBUG_DISABLE_DISSECTORS=true
run_single_bench "05" "no_dissectors" || exit 1
export KUBESHARK_DEBUG_DISABLE_DISSECTORS=false

export KUBESHARK_DEBUG_DISABLE_EMITTING=true
run_single_bench "06" "no_emit" || exit 1
export KUBESHARK_DEBUG_DISABLE_EMITTING=false

export KUBESHARK_DEBUG_DISABLE_SENDING=true
run_single_bench "07" "no_send" || exit 1
export KUBESHARK_DEBUG_DISABLE_SENDING=false

run_single_bench "08" "normal" || exit 1

generate_bench_graph || exit 1
log "Output written to to $KUBESHARK_BENCHMARK_OUTPUT_DIR"
