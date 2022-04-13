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

        cy.get('#list [id^=entry]').each(entryElement => {
            entryElement.click();
            cy.get('#tbody-Headers').should('be.visible');
            isValueExistsInElement(false, 'Ignored-User-Agent', '#tbody-Headers');
        });
    });
}

