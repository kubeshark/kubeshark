import {findLineAndCheck, getExpectedDetailsDict} from "../testHelpers/StatusBarHelper";

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

it('top bar check', function () {
    const podName1 = 'httpbin', namespace1 = 'mizu-tests';
    const podName2 = 'httpbin2', namespace2 = 'mizu-tests';

    cy.get('.podsCount').trigger('mouseover');
    findLineAndCheck(getExpectedDetailsDict(podName1, namespace1));
    findLineAndCheck(getExpectedDetailsDict(podName2, namespace2))
});

// it('pause straming button should send specific request', function () {
//     cy.get('#total-entries').then(entries => {
//         const entriesNum = entries.text();
//         cy.intercept('http://localhost:8899/entries/49?query=undefined');
//     })
// });

it('on mouse hover, the plus icon should be visible', function () {
    cy.reload()
    cy.get('#entry-48 > :nth-child(1) > .Protocol_base__1EqkO').trigger('mouseover')
    cy.get('.Queryable-Tooltip').should('be.visible')
}); 
