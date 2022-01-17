import {check} from '../test-helpers/TrafficHelper';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
})

check(true, Cypress.env('redactHeaderContent'), '#tbody-Headers');
check(true, Cypress.env('redactBodyContent'), '.hljs');
