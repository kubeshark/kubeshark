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
WEBSOCKET_TIMEOUT = 5
ENTRIES_ENDPOINT = 'http://%s:%s/entries' % (HOST, PORT)
WEBSOCKET_ENDPOINT = 'ws://%s:%s/ws' % (HOST, PORT)

dir_path = os.path.dirname(os.path.realpath(__file__))


class Entry:

    def __init__(self, _id: int, base: dict) -> None:
        self.id = _id # type: int
        self.base = base # type: dict
        self.data = {} # type: dict

    def set_data(self, data: dict) -> None:
        self.data = data

class DataGroup:

    def __init__(self, query: str) -> None:
        self.query = query # type: str
        self.entries = [] # type: List[Union[Entry, str]]


class Suite:

    def __init__(self) -> None:
        self.data_groups = [] # type: List[DataGroup]


queries = [
    '',
    # 'method == "PING"',
    # 'method == "FLUSHDB"',
]

current_data_group = None # type: List[DataGroup]

def on_message(ws, message):
    data = json.loads(message)
    if data['messageType'] == 'entry':
        data['data'].pop('isOutgoing', None)
        entry = Entry(_id=data['data']['id'], base=data)
        current_data_group.entries.append(entry)

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
    for query in queries:
        print('Running query "%s"...' % query)
        data_group = DataGroup(query=query)
        suite.data_groups.append(data_group)
        current_data_group = data_group
        on_open_extended = partial(on_open, query=query)
        ws = websocket.WebSocketApp(
            WEBSOCKET_ENDPOINT,
            on_open=on_open_extended,
            on_message=on_message,
            on_error=on_error,
            on_close=on_close
        )
        ws.run_forever()

        print('Streamed %d entries.' % len(data_group.entries))

    for data_group in suite.data_groups:
        print('[Query: "%s"] Fetching full entries...' % data_group.query)
        for i, entry in enumerate(data_group.entries):
            # print("Fetch progress: %d/%d \r" % (i, len(data_group.entries)), end='\r')
            url = '%s/%d' % (ENTRIES_ENDPOINT, entry.id)
            resp = requests.get(url=url, params={'query': data_group.query})
            data = resp.json()
            entry.set_data(data)

    for data_group in suite.data_groups:
        entries = []
        data_group.entries = sorted(data_group.entries, key=lambda x: (x.base['data']['timestamp']))
        for i, entry in enumerate(data_group.entries):
            entry.base['data']['id'] = 0
            entries.append(jsonpickle.encode(entry.base))
        data_group.entries = entries

    if len(sys.argv) > 1 and sys.argv[1] == "update":
        serialized = jsonpickle.encode(suite)
        suite_path = '%s/suite.json' % dir_path

        with open('%s/suite.json' % dir_path, 'w') as outfile:
            outfile.write(json.dumps(json.loads(serialized), indent=2))

        print("The test suite is saved into: %s" % suite_path)
