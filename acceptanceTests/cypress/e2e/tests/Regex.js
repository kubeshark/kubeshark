import {getExpectedDetailsDict, checkLine} from '../testHelpers/StatusBarHelper';


it('opening', function () {
    cy.visit(Cypress.env('testUrl'));
    cy.get(`[data-cy="podsCountText"]`).trigger('mouseover');

    cy.get('[data-cy="expandedStatusBar"] > :nth-child(2) > > :nth-child(2) >').should('have.length', 1); // one line

    checkLine(1, getExpectedDetailsDict(Cypress.env('name'), Cypress.env('namespace')));
});
