import {findLineAndCheck, getExpectedDetailsDict} from "../testHelpers/StatusBarHelper";
import {verifyMinimumEntries} from "../testHelpers/TrafficHelper";

it('check', function () {
    const podName = Cypress.env('name'), namespace = Cypress.env('namespace');
    const port = Cypress.env('port');
    cy.intercept('GET', `http://localhost:${port}/status/tap`).as('statusTap');

    cy.visit(`http://localhost:${port}`);
    cy.wait('@statusTap').its('response.statusCode').should('match', /^2\d{2}/);

    verifyMinimumEntries();

    cy.get('.podsCount').trigger('mouseover');
    findLineAndCheck(getExpectedDetailsDict(podName, namespace));
});
