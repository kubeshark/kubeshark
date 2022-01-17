import {isValueExistsInElement, isValueExistsInElement} from '../test-helpers/TrafficHelper';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
})

isValueExistsInElement(false, Cypress.env('redactHeaderContent'), '#tbody-Headers');
isValueExistsInElement(false, Cypress.env('redactBodyContent'), '.hljs');
