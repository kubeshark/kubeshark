#!/usr/bin/python3
# -*- coding: utf-8 -*-

"""
.. module:: test
    :synopsis: Automation for the PCAP testbed
"""

import _thread
import time
import json
import sys
import os
from functools import partial
from typing import List, Union

import websocket
import jsonpickle
import requests


HOST = 'localhost'
PORT = '8899'
WEBSOCKET_TIMEOUT = 0.01
ENTRIES_ENDPOINT = 'http://%s:%s/entries' % (HOST, PORT)
WEBSOCKET_ENDPOINT = 'ws://%s:%s/ws' % (HOST, PORT)
TOLERANCE = 10

dir_path = os.path.dirname(os.path.realpath(__file__))


class Query:

    def __init__(self, query: str, consistent: bool) -> None:
        self.query = query # type: str
        self.number_of_records = 0 # type: int
        self.consistent = consistent # type: bool
        self.ids = [] # type: List[int]


queries = [
    Query('', False),
    Query('amqp', True),
    Query('amqp and method == "connection start"', True),
    Query('amqp and method == "connection close" and request.replyText == "kthxbai"', True),
    Query('amqp and method == "connection start" and request.versionMajor == "0" and request.versionMinor == "9"', True),
    Query('amqp and method == "queue declare" and request.queue == "test-integration-declared-passive-queue"', True),
    Query('amqp and request.redelivered == "false" and request.routingKey == "test-corrupted-message-regression"', True),
    Query('amqp and timestamp >= datetime("02/03/2022, 9:02:18.034 AM") and timestamp <= datetime("02/03/2022, 9:02:32.634 AM")', True),
    Query('amqp and src.port == "49346" and dst.port == "5672"', True),
    Query('amqp and method == "basic publish" and summary == "not-existing-exchange" and src.port == "49332" and dst.port == "5672" and timestamp >= datetime("02/03/2022, 9:02:17.935 AM")', True),
    Query('amqp and method == "basic publish" and summary == "not-existing-exchange" and request.routingKey == "some-key"', True),
    Query('amqp and method == "queue bind" and summary == "test-basic-ops-queue"', True),
    Query('amqp and method == "exchange declare" and summary == "test-integration-declared-passive-exchange" and request.autoDelete == "true" and request.durable == "false" and request.exchange == "test-integration-declared-passive-exchange" and request.internal == "false" and request.noWait == "false" and request.passive == "true" and request.type == "direct"', True),
    Query('amqp and method == "queue declare" and summary == "test-integration-declared-passive-queue" and request.autoDelete == "true" and request.durable == "true" and request.exclusive == "false" and request.noWait == "false" and request.queue == "true" and request.queue == "test-integration-declared-passive-queue"', True),
    Query('amqp and method == "connection close" and summary == "NOT_IMPLEMENTED - active=false" and request.classId == "20" and request.methodId == "20" and request.replyCode == "" and request.replyText == "NOT_IMPLEMENTED - active=false"', True),
    Query('amqp and method == "basic consume" and summary == "test.integration.consumer-flow" and timestamp >= datetime("02/03/2022, 9:02:01.291 AM") and request.consumerTag == "ctag-/tmp/go-build1228491040/b001/amqp.test-7" and request.queue == "test.integration.consumer-flow"', True),
    Query('grpc', False),
    Query('http', False),
    Query('http2', False),
    Query('kafka', False),
    Query('kafka and request.apiKey == "Produce" and request.apiVersion == "8" and request.correlationID == "2" and request.payload.topicData.partitions.partitionData.records.recordBatch["firstTimestamp"] == 1643962317412', True),
    Query('kafka and request.apiKey == "Fetch" and request.apiVersion == "11" and request.size == "104" and request.payload.maxBytes == "262144" and request.payload.topics[0].partitions[0]["fetchOffset"] == 640', True),
    Query('kafka and request.apiKey == "ApiVersions" and request.apiVersion == "0" and request.correlationID == "1"', True),
    Query('kafka and request.apiKey == "ListOffsets" and request.apiVersion == "1" and summary == "kafka-go-4619ebd3231b3901" and response.correlationID == "3"', True),
    Query('kafka and request.apiKey == "Metadata" and request.apiVersion == "1" and request.correlationID == "2" and request.size == "94" and response.payload.brokers == "[{\"host\":\"localhost\",\"nodeId\":0,\"port\":9092,\"rack\":\"\"}]" and response.payload.controllerID == "0"', True),
    Query('kafka and request.apiKey == "Produce" and request.apiVersion == "7" and request.clientID == "kafka-go.test@Corsair (github.com/segmentio/kafka-go)" and request.correlationID == "13" and request.size == "183" and request.payload.timeout == "2147483647" and request.payload.transactionalID == 1 and request.payload.topicData.partitions.partitionData.records.recordBatch["firstTimestamp"] == 1643962320535 and request.payload.topicData.partitions.partitionData.records.recordBatch.record[0]["attributes"] == 0 and request.payload.topicData.partitions.partitionData.records.recordBatch.record[0].value == "6"', True),
    Query('kafka and request.apiKey == "Fetch" and request.apiVersion == "10" and summary == "kafka-go-260b77aad1937b0f" and response.payload.responses[0].partitionResponses[0].recordSet.recordBatch.record[0].value == "3"', True),
    Query('redis', False),
    Query('redis and method == "PING"', False),
    Query('redis and method == "FLUSHDB"', False),
    Query('redis and request.command == "GET" and request.key == "counter3"', True),
    Query('redis and request.command == "MULTI" and request.type == "Array"', False),
    Query('redis and request.command == "SUBSCRIBE" and request.key == "mychannel1"', True),
    Query('redis and request.command == "PING" and request.type == "Array"', False),
    Query('redis and request.command == "GET" and request.key == "A" and request.type == "Array"', True),
    Query('redis and request.command == "DEL" and request.key == "A" and request.type == "Array"', True),
    Query('redis and request.command == "SET" and request.key == "key6" and request.type == "Array" and request.value == "value"', False),
    Query('redis and request.command == "INFO" and request.key == "keyspace" and request.type == "Array" and response.type == "Bulk String" and response.value == "# Keyspace"', False),
    Query('redis and response.keyword == "OK" and response.type == "Simple String"', False),
    Query('redis and request.command == "EVALSHA" and request.key == "b3b5be469962cc72e488ee86a39ed8b552e3ed35" and request.type == "Array" and request.value == "[1, key2, value]" and response.type == "Error" and response.value == "NoScriptError: NOSCRIPT No matching script. Please use EVAL."', False),
    Query('redis and request.command == "EVAL" and request.key == "\n\t\t\t\tlocal r = redis.call(\'SET\', KEYS[1], ARGV[1])\n\t\t\t\treturn r\n\t\t\t" and request.type == "Array" and response.keyword == "OK" and response.type == "Simple String"', True),
    Query('redis and request.command == "SCRIPT" and request.key == "flush" and request.type == "Array"', True),
    Query('redis and request.command == "SCRIPT" and request.key == "load" and request.type == "Array" and request.value == "return \'Unique script\'" and response.value == "586deab7f5d7baecfdab4753abeff059e87bebe0"', True),
    Query('redis and request.command == "EXPIRE" and request.key == "D" and request.type == "Array" and request.value == "14400" and response.type == "Integer" and response.value == "1"', True),
    Query('redis and (request.command == "DBSIZE" or request.command == "TTL" or request.command == "WATCH" or request.command == "UNWATCH")', False),
    Query('redis and request.command == "WATCH" and request.key == "{shard}key1" and request.value == "{shard}key2"', True),
    Query('redis and request.command == "SET" and request.key == ";��Q&���_�Z7�\u001eω;�;���sh��\u0019\u0014���\u0001?��.x����W�;kE7\n!)\u001d�z7��߯�Qe\u0016N�"', True),
    Query('redis and request.command == "AUTH" and request.key == "password1" and response.type == "Error" and response.value == "DataError: ERR AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?"', True),
    Query('redis and request.command == "CLIENT" and request.key == "setname" and request.type == "Array" and request.value == "foobar"', True),
    Query('redis and request.command == "CLIENT" and request.key == "getname" and request.type == "Array" and response.type == "Bulk String" and response.value == "foobar"', True),
] # type: List[Query]


class Suite:

    def __init__(self) -> None:
        self.queries = [] # type: List[DataGroup]


def on_message(ws, message, query_obj):
    data = json.loads(message)
    if data['messageType'] == 'entry':
        query_obj.ids.append(data['data']['id'])
    elif data['messageType'] == 'queryMetadata':
        if data['data']['leftOff'] == data['data']['total']:
            def run(*args):
                time.sleep(WEBSOCKET_TIMEOUT)
                ws.close()
            _thread.start_new_thread(run, ())

def on_error(ws, error):
    print(error)

def on_close(ws, close_status_code, close_msg):
    print('WebSocket is closed.')

def on_open(ws, query):
    ws.send(query) # query


if __name__ == "__main__":
    suite = Suite()
    # websocket.enableTrace(True)

    # Run all the queries
    for q in queries:
        print('Running query "%s"...' % q.query)
        suite.queries.append(q)
        on_open_extended = partial(on_open, query=q.query)
        on_message = partial(on_message, query_obj=q)
        ws = websocket.WebSocketApp(
            WEBSOCKET_ENDPOINT,
            on_open=on_open_extended,
            on_message=on_message,
            on_error=on_error,
            on_close=on_close
        )
        ws.run_forever()

        q.number_of_records = len(q.ids)

        print('Streamed %d entries.' % q.number_of_records)

    suite_path = '%s/suite.json' % dir_path
    if len(sys.argv) > 1 and sys.argv[1] == "update":
        for q in suite.queries:
            del q.ids
        serialized = jsonpickle.encode(suite)

        with open(suite_path, 'w') as outfile:
            outfile.write(json.dumps(json.loads(serialized), indent=2) + '\n')

        print("The test suite is saved into: %s" % suite_path)
    else:
        # Fetch the full data of all of the entries in the first query (for smoke testing)
        q = suite.queries[0]
        print('[Query: "%s"] Fetching full entries...' % q.query)
        for _id in q.ids:
            # print("Fetch progress: %d/%d \r" % (i, len(data_group.entries)), end='\r')
            url = '%s/%d' % (ENTRIES_ENDPOINT, _id)
            resp = requests.get(url=url, params={'query': q.query})
            assert resp.status_code == 200
            data = resp.json()
            assert data

        for q in suite.queries:
            del q.ids

        with open(suite_path, 'r') as infile:
            suite_ref = jsonpickle.decode(infile.read())

        for i, query_ref in enumerate(suite_ref.queries):
            if i >= len(suite.queries):
                print("Query length out of range!")
                assert False

            query = suite.queries[i]
            print('c:', query.__dict__)
            print('r:', query_ref.__dict__)
            if query_ref.consistent:
                assert query.number_of_records == query_ref.number_of_records
            else:
                sigma = TOLERANCE / 100 * query_ref.number_of_records
                lower_limit = query_ref.number_of_records - sigma
                upper_limit = query_ref.number_of_records + sigma
                assert lower_limit <= query.number_of_records and query.number_of_records <= upper_limit
