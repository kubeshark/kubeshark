import {isValueExistsInElement} from '../testHelpers/TrafficHelper';

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

isValueExistsInElement(false, Cypress.env('redactHeaderContent'), '#tbody-Headers');
isValueExistsInElement(false, Cypress.env('redactBodyContent'), Cypress.env('bodyJsonClass'));
