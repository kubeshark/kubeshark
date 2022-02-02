import {isValueExistsInElement, verifyAtLeastXentries} from '../testHelpers/TrafficHelper';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

verifyAtLeastXentries();

isValueExistsInElement(true, Cypress.env('redactHeaderContent'), '#tbody-Headers');
isValueExistsInElement(true, Cypress.env('redactBodyContent'), '.hljs');
