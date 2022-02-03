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
WEBSOCKET_TIMEOUT = 3
ENTRIES_ENDPOINT = 'http://%s:%s/entries' % (HOST, PORT)
WEBSOCKET_ENDPOINT = 'ws://%s:%s/ws' % (HOST, PORT)

queries = [
    ('', False),
    ('amqp', True),
    ('method == "connection start"', True),
    ('method == "connection close"', True),
    ('method == "connection start" and request.versionMajor == "0" and request.versionMinor == "9"', True),
    ('method == "queue declare" and request.queue == "test-integration-declared-passive-queue"', True),
    ('redis', False),
    ('redis and method == "PING"', False),
    ('redis and method == "FLUSHDB"', False),
    ('request.command == "GET" and request.key == "counter3"', False),
    ('request.command == "MULTI" and request.type == "Array"', False),
    ('request.command == "SUBSCRIBE" and request.key == "mychannel1"', False),
]


dir_path = os.path.dirname(os.path.realpath(__file__))


class Query:

    def __init__(self, query: str) -> None:
        self.query = query # type: str
        self.number_of_records = 0
        self.consistent = False
        self.ids = [] # type: List[int]


class Suite:

    def __init__(self) -> None:
        self.queries = [] # type: List[DataGroup]


current_query = None # type: Union[None, Query]

def on_message(ws, message):
    data = json.loads(message)
    if data['messageType'] == 'entry':
        data['data'].pop('isOutgoing', None)
        current_query.ids.append(data['data']['id'])

def on_error(ws, error):
    print(error)

def on_close(ws, close_status_code, close_msg):
    print('WebSocket is closed.')

def on_open(ws, query):
    def run(*args):
        ws.send(query) # query
        time.sleep(WEBSOCKET_TIMEOUT)
        ws.close()
        print('thread terminating...')
    _thread.start_new_thread(run, ())


if __name__ == "__main__":
    suite = Suite()
    # websocket.enableTrace(True)
    for query, consistent in queries:
        print('Running query "%s"...' % query)
        q = Query(query=query)
        q.consistent = consistent
        suite.queries.append(q)
        current_query = q
        on_open_extended = partial(on_open, query=query)
        ws = websocket.WebSocketApp(
            WEBSOCKET_ENDPOINT,
            on_open=on_open_extended,
            on_message=on_message,
            on_error=on_error,
            on_close=on_close
        )
        ws.run_forever()

        q.number_of_records = len(q.ids)
        del q.ids

        print('Streamed %d entries.' % q.number_of_records)

    # for q in suite.queries:
    #     print('[Query: "%s"] Fetching full entries...' % q.query)
    #     for _id in q.ids:
    #         # print("Fetch progress: %d/%d \r" % (i, len(data_group.entries)), end='\r')
    #         url = '%s/%d' % (ENTRIES_ENDPOINT, _id)
    #         resp = requests.get(url=url, params={'query': q.query})
    #         data = resp.json()

    if len(sys.argv) > 1 and sys.argv[1] == "update":
        serialized = jsonpickle.encode(suite)
        suite_path = '%s/suite.json' % dir_path

        with open('%s/suite.json' % dir_path, 'w') as outfile:
            outfile.write(json.dumps(json.loads(serialized), indent=2))

        print("The test suite is saved into: %s" % suite_path)
