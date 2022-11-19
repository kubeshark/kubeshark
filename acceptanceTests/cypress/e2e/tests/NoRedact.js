import {isValueExistsInElement} from '../testHelpers/TrafficHelper';

it('Loading Kubeshark', function () {
    cy.visit(Cypress.env('testUrl'));
});

isValueExistsInElement(false, Cypress.env('redactHeaderContent'), '#tbody-Headers');
isValueExistsInElement(false, Cypress.env('redactBodyContent'), Cypress.env('bodyJsonClass'));
