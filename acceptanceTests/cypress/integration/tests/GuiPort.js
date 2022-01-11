it('check', function () {
    cy.visit(`http://localhost:${Cypress.env('port')}/`);

    cy.get('.header').should('be.visible');
    cy.get('.TrafficPageHeader').should('be.visible');
    cy.get('.TrafficPage-ListContainer').should('be.visible');
    cy.get('.TrafficPage-Container').should('be.visible');
});
