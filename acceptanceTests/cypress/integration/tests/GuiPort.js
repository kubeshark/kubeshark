it('check', function () {
    // cy.visit(`http://localhost:${Cypress.env('port')}/`)
    cy.visit(`http://localhost:8899/`)

    cy.get('.header').should('be.visible')
    cy.get('.TrafficPageHeader').should('be.visible')
    cy.get('.TrafficPage-ListContainer').should('be.visible')
    cy.get('.TrafficPage-Container').should('be.visible')
})
