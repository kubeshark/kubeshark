import {checkFilterByMethod, valueTabs,} from "../testHelpers/TrafficHelper";

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

const redisProtocolDetails = {name: 'redis', text: 'Redis Serialization Protocol'};

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'PING',
    summary: null,
    value: null
})

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'SET',
    summary: 'key',
    value: {tab: valueTabs.request, regex: /^\[value, keepttl]$/mg}
})

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'EXISTS',
    summary: 'key',
    value: {tab: valueTabs.response, regex: /^1$/mg}
})

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'GET',
    summary: 'key',
    value: {tab: valueTabs.response, regex: /^value$/mg}
})

checkFilterByMethod({
    protocol: redisProtocolDetails,
    method: 'DEL',
    summary: 'key',
    value: {tab: valueTabs.response, regex: /^1$|^0$/mg}
})
