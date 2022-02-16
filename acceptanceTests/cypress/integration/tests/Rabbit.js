import {checkFilterByMethod, valueTabs,} from "../testHelpers/TrafficHelper";

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
    cy.get('#total-entries').invoke('text').should('match', /^[4-7][0-9]$/m)
});

const rabbitProtocolDetails = {name: 'AMQP', text: 'Advanced Message Queuing Protocol 0-9-1'};

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method: 'exchange declare',
    summary: 'exchange',
    value: null
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method: 'queue declare',
    summary: 'queue',
    value: null
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method:  'queue bind',
    summary: 'queue',
    value: null
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method: 'basic publish',
    summary: 'exchange',
    value: {tab: valueTabs.request, regex: /^message$/mg}
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method:  'basic consume',
    summary: 'queue',
    value: null
});

checkFilterByMethod({
    protocol: rabbitProtocolDetails,
    method:  'basic deliver',
    summary: 'exchange',
    value: {tab: valueTabs.request, regex: /^message$/mg}
});
