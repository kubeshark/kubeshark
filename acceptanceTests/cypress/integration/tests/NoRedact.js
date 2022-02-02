import {isValueExistsInElement, verifyAtLeastXentries} from '../testHelpers/TrafficHelper';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

verifyAtLeastXentries();

isValueExistsInElement(false, Cypress.env('redactHeaderContent'), '#tbody-Headers');
isValueExistsInElement(false, Cypress.env('redactBodyContent'), '.hljs');
