import {findLineAndCheck, getExpectedDetailsDict} from "../testHelpers/StatusBarHelper";
import {verifyMinimumEntries} from "../testHelpers/TrafficHelper";

it('check', function () {
    const podName = Cypress.env('name'), namespace = Cypress.env('namespace');
    cy.intercept('GET', `http://localhost:${Cypress.env('port')}/status/tap`).as('statusTap');

    cy.visit(`http://localhost:${Cypress.env('port')}`);
    cy.wait('@statusTap').its('response.statusCode').should('match', /^2\d{2}/)

    verifyMinimumEntries();

    cy.get('.podsCount').trigger('mouseover');
    findLineAndCheck(getExpectedDetailsDict(podName, namespace));
});
