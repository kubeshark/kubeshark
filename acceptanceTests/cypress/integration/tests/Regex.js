import StatusBarFunctions from "../page_objects/StatusBar";
const base = new StatusBarFunctions()

it('opening', function () {
    cy.visit(Cypress.env('testUrl'))
    cy.get('.podsCount').trigger('mouseover')

    cy.get('.expandedStatusBar > :nth-child(2) > > :nth-child(2) >').should('have.length', 1) // one line

    base.checkLine(1, base.getExpectedDetailsDict(Cypress.env('name'), Cypress.env('namespace')))
});
