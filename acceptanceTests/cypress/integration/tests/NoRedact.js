import {check} from '../page_objects/Traffic';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
})

check(false, Cypress.env('redactHeaderContent'), '#tbody-Headers');
check(false, Cypress.env('redactBodyContent'), '.hljs');
