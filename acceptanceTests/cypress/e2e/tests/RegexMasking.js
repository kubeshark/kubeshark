import {isValueExistsInElement} from "../testHelpers/TrafficHelper";

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

isValueExistsInElement(true, Cypress.env('regexMaskingBodyContent'), Cypress.env('bodyJsonClass'));
