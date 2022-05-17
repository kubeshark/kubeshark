#!/bin/bash

[ -z "$MIZU_HOME" ] && { echo "MIZU_HOME is missing"; exit 1; }
[ -z "$MIZU_BENCHMARK_OUTPUT_DIR" ] && export MIZU_BENCHMARK_OUTPUT_DIR="/tmp/mizu-benchmark-results-$(date +%d-%m-%H-%M)"
[ -z "$MIZU_BENCHMARK_CLIENT_PERIOD" ] && export MIZU_BENCHMARK_CLIENT_PERIOD="1m"
[ -z "$MIZU_BENCHMARK_URL" ] && export MIZU_BENCHMARK_URL="http://localhost:8081/data/b.1000.json"
[ -z "$MIZU_BENCHMARK_RUN_COUNT" ] && export MIZU_BENCHMARK_RUN_COUNT="3"
[ -z "$MIZU_BENCHMARK_QPS" ] && export MIZU_BENCHMARK_QPS="500"
[ -z "$MIZU_BENCHMARK_CLIENTS_COUNT" ] && export MIZU_BENCHMARK_CLIENTS_COUNT="5"

function log() {
	local message=$@
	printf "[%s] %s\n" "$(date "+%d-%m %H:%M:%S")" "$message"
}

function run_single_bench() {
	local mode_num=$1
	local mode_str=$2

	log "Starting ${mode_num}_${mode_str} (runs: $MIZU_BENCHMARK_RUN_COUNT) (period: $MIZU_BENCHMARK_CLIENT_PERIOD)"

	for ((i=0;i<"$MIZU_BENCHMARK_RUN_COUNT";i++)); do
		log "  $i: Running tapper"
		rm -f tapper.log
		nohup ./agent/build/mizuagent --tap --api-server-address ws://localhost:8899/wsTapper -i lo -stats 10 > tapper.log 2>&1 &

		log "  $i: Running client (hey)"
		hey -z $MIZU_BENCHMARK_CLIENT_PERIOD -c MIZU_BENCHMARK_CLIENTS_COUNT -q $MIZU_BENCHMARK_QPS $MIZU_BENCHMARK_URL > /dev/null || return 1

		log "  $i: Killing tapper"
		kill -9 $(ps -ef | grep agent/build/mizuagent | grep tap | grep -v grep | awk '{ print $2 }') > /dev/null 2>&1

		local output_file=$MIZU_BENCHMARK_OUTPUT_DIR/${mode_num}_${mode_str}_${i}.log
		log "  $i: Copying output to $output_file"
		cp tapper.log $output_file || return 1
	done
}

function generate_bench_graph() {
	cd performance_analysis/ || return 1
	source venv/bin/activate
	python plot_from_tapper_logs.py $MIZU_BENCHMARK_OUTPUT_DIR/* || return 1
	mv graph.png $MIZU_BENCHMARK_OUTPUT_DIR || return 1
}

mkdir -p $MIZU_BENCHMARK_OUTPUT_DIR
rm -f $MIZU_BENCHMARK_OUTPUT_DIR/*
log "Writing output to $MIZU_BENCHMARK_OUTPUT_DIR"

cd $MIZU_HOME || exit 1

export HOST_MODE=0
export MIZU_TAPPER_NO_PCAP=false
export MIZU_TAPPER_NO_ASSEMBLER=false
export MIZU_TAPPER_NO_TAP_TARGET=false
export MIZU_TAPPER_NO_OTHER_EXTENSSIONS=false
export MIZU_TAPPER_NO_DISSECTORS=false
export MIZU_TAPPER_NO_EMITTER=false
export MIZU_TAPPER_NO_SENDING=false

export MIZU_TAPPER_NO_PCAP=true
run_single_bench "1" "no_pcap" || exit 1
export MIZU_TAPPER_NO_PCAP=false

export MIZU_TAPPER_NO_ASSEMBLER=true
run_single_bench "2" "no_assembler" || exit 1
export MIZU_TAPPER_NO_ASSEMBLER=false

export MIZU_TAPPER_NO_TAP_TARGET=true
run_single_bench "3" "no_tap_targets" || exit 1
export MIZU_TAPPER_NO_TAP_TARGET=false

export MIZU_TAPPER_NO_OTHER_EXTENSSIONS=true
run_single_bench "4" "only_http" || exit 1
export MIZU_TAPPER_NO_OTHER_EXTENSSIONS=false

export MIZU_TAPPER_NO_DISSECTORS=true
run_single_bench "5" "no_dissectors" || exit 1
export MIZU_TAPPER_NO_DISSECTORS=false

export MIZU_TAPPER_NO_EMITTER=true
run_single_bench "6" "no_emit" || exit 1
export MIZU_TAPPER_NO_EMITTER=false

export MIZU_TAPPER_NO_SENDING=true
run_single_bench "7" "no_send" || exit 1
export MIZU_TAPPER_NO_SENDING=false

run_single_bench "8" "normal" || exit 1

generate_bench_graph || exit 1
log "Output written to to $MIZU_BENCHMARK_OUTPUT_DIR"
