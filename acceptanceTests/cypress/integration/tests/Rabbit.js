import {checkFilterByMethod, valueTabs,} from "../testHelpers/TrafficHelper";

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

const rabbitProtocolDetails = {name: 'AMQP', text: 'Advanced Message Queuing Protocol 0-9-1'};

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method: 'exchange declare',
    methodQuery: 'request.method == "exchange declare"',
    summary: 'exchange',
    summaryQuery: 'request.exchange == "exchange"',
    value: null
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method: 'queue declare',
    methodQuery: 'request.method == "queue declare"',
    summary: 'queue',
    summaryQuery: 'request.queue == "queue"',
    value: null
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method:  'queue bind',
    methodQuery: 'request.method == "queue bind"',
    summary: 'queue',
    summaryQuery: 'request.queue == "queue"',
    value: null
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method: 'basic publish',
    methodQuery: 'request.method == "basic publish"',
    summary: 'exchange',
    summaryQuery: 'request.exchange == "exchange"',
    value: {tab: valueTabs.request, regex: /^message$/mg}
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method:  'basic consume',
    methodQuery: 'request.method == "basic consume"',
    summary: 'queue',
    summaryQuery: 'request.queue == "queue"',
    value: null
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method:  'basic deliver',
    methodQuery: 'request.method == "basic deliver"',
    summary: 'exchange',
    summaryQuery: 'request.queue == "exchange"',
    value: {tab: valueTabs.request, regex: /^message$/mg}
});
