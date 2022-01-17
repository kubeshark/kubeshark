import {check} from '../test-helpers/TrafficHelper';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
})

check(false, Cypress.env('redactHeaderContent'), '#tbody-Headers');
check(false, Cypress.env('redactBodyContent'), '.hljs');
