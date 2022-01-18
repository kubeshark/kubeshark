import {isValueExistsInElement} from "../testHelpers/TrafficHelper";

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

it('going through each entry', function () {

    cy.get('#total-entries').then(number => {
        const getNum = () => {
            const numOfEntries = number.text();
            return parseInt(numOfEntries);
        };
        cy.wrap({ there: getNum }).invoke('there').should('be.gte', 35);
        checkThatAllEntriesShown();
        const entriesNum = getNum();
        [...Array(entriesNum).keys()].map(checkEntry);
    });
});

function checkThatAllEntriesShown() {
    cy.get('#entries-length').then(number => {
        if (number.text() === '1')
            cy.get('[title="Fetch old records"]').click();
    });
}

function checkEntry(entryIndex) {
    cy.get(`#entry-${entryIndex}`).click();
    cy.get('#tbody-Headers').should('be.visible');
    isValueExistsInElement(false, 'Ignored-User-Agent', '#tbody-Headers');
}
