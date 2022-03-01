import {isValueExistsInElement} from '../testHelpers/TrafficHelper';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

isValueExistsInElement(true, Cypress.env('redactHeaderContent'), '#tbody-Headers');
isValueExistsInElement(true, Cypress.env('redactBodyContent'), Cypress.env('bodyJsonClass'));
