it('first test', function () {
    cy.visit('http://localhost:8899/')

    cy.get('.title > img').should('be.visible')
    cy.get('.description').invoke('text').should('eq', 'Traffic viewer for Kubernetes')
})

