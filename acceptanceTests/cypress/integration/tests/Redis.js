import {checkFilterByMethod, valueTabs,} from "../testHelpers/TrafficHelper";

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

const redisProtocolDetails = {name: 'redis', text: 'Redis Serialization Protocol'};
const numberOfRecords = 5;

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'PING',
    methodQuery: 'request.command == "PING"',
    summary: null,
    summaryQuery: '',
    numberOfRecords: numberOfRecords,
    value: null
})

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'SET',
    methodQuery: 'request.command == "SET"',
    summary: 'key',
    summaryQuery: 'request.key == "key"',
    numberOfRecords: numberOfRecords,
    value: {tab: valueTabs.request, regex: /^\[value, keepttl]$/mg}
})

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'EXISTS',
    methodQuery: 'request.command == "EXISTS"',
    summary: 'key',
    summaryQuery: 'request.key == "key"',
    numberOfRecords: numberOfRecords,
    value: {tab: valueTabs.response, regex: /^1$/mg}
})

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'GET',
    methodQuery: 'request.command == "GET"',
    summary: 'key',
    summaryQuery: 'request.key == "key"',
    numberOfRecords: numberOfRecords,
    value: {tab: valueTabs.response, regex: /^value$/mg}
})

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'DEL',
    methodQuery: 'request.command == "DEL"',
    summary: 'key',
    summaryQuery: 'request.key == "key"',
    numberOfRecords: numberOfRecords,
    value: {tab: valueTabs.response, regex: /^1$|^0$/mg}
})
