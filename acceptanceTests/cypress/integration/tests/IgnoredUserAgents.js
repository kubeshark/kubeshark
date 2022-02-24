import {
    checkThatAllEntriesShown,
    isValueExistsInElement,
    resizeToHugeMizu,
} from "../testHelpers/TrafficHelper";

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

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

function checkEntry(entryIndex) {
    cy.get(`#entry-${entryIndex}`).click();
    cy.get('#tbody-Headers').should('be.visible');
    isValueExistsInElement(false, 'Ignored-User-Agent', '#tbody-Headers');
}
