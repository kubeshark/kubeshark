import {findLineAndCheck, getExpectedDetailsDict} from "../testHelpers/StatusBarHelper";
import {verifyAtLeastXentries} from "../testHelpers/TrafficHelper";

it('check', function () {
    const podName = 'httpbin', namespace = 'mizu-tests';
    cy.intercept('GET', 'http://localhost:8898/status/tap').as('statusTap');

    cy.visit(`http://localhost:8898`);
    cy.wait('@statusTap').its('response.statusCode').should('match', /^2\d{2}/)

    verifyAtLeastXentries();
    
    cy.get('.podsCount').trigger('mouseover');
    findLineAndCheck(getExpectedDetailsDict(podName, namespace));

    cy.get('.header').should('be.visible');
    cy.get('.TrafficPageHeader').should('be.visible');
    cy.get('.TrafficPage-ListContainer').should('be.visible');
    cy.get('.TrafficPage-Container').should('be.visible');
});
