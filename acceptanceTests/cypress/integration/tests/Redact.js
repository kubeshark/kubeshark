import {isValueExistsInElement, isValueExistsInElement} from '../test-helpers/TrafficHelper';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
})

isValueExistsInElement(true, Cypress.env('redactHeaderContent'), '#tbody-Headers');
isValueExistsInElement(true, Cypress.env('redactBodyContent'), '.hljs');
