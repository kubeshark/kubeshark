it('should', function () {
    cy.visit('http://localhost:8899/')

    cy.get('.connectionText', {timeout : 20000}).should('have.text', 'streaming paused')
    cy.get('.styles_scrollable-div__prSCv').scrollTo(0, -300)
})
