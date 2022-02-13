import {isValueExistsInElement, verifyMinimumEntries} from "../testHelpers/TrafficHelper";

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

verifyMinimumEntries();

isValueExistsInElement(true, Cypress.env('regexMaskingBodyContent'), Cypress.env('bodyJsonClass'));
