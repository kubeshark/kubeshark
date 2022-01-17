import {isValueExistsInElement} from "../testHelpers/TrafficHelper";
it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
})

it('should ', function () {
    cy.get('#total-entries').next().then(bottomLine => {
        const number = bottomLine.text();
        cy.log(number)

    })
});
