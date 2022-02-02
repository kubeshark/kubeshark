import {isValueExistsInElement, resizeToHugeMizu, verifyAtLeastXentries} from "../testHelpers/TrafficHelper";

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

verifyAtLeastXentries();

checkEntries();

function checkEntries() {
    it('checking all entries', function () {
        checkThatAllEntriesShown();
        resizeToHugeMizu();

        cy.get('#total-entries').then(number => {
            const numOfEntries = parseInt(number.text());
            [...Array(numOfEntries).keys()].map(checkEntry);
        });
    });
}

function checkThatAllEntriesShown() {
    cy.get('#entries-length').then(number => {
        if (number.text() === '1')
            cy.log('clicked the fetch old records butotn');
            cy.get('[title="Fetch old records"]').click();
    });
}

function checkEntry(entryIndex) {
    cy.get(`#entry-${entryIndex}`).click();
    cy.get('#tbody-Headers').should('be.visible');
    isValueExistsInElement(false, 'Ignored-User-Agent', '#tbody-Headers');
}
