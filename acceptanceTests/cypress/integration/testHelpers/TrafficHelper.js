export function isValueExistsInElement(shouldInclude, content, domPathToContainer){
    it(`should ${shouldInclude ? '' : 'not'} include '${content}'`, function () {
        cy.get(domPathToContainer).then(htmlText => {
            const allTextString = htmlText.text();
            if (allTextString.includes(content) !== shouldInclude)
                throw new Error(`One of the containers part contains ${content}`)
        });
    });
}

export function resizeToHugeMizu() {
    cy.viewport(1920, 3500);
}

export function resizeToNormalMizu() {
    cy.viewport(1920, 1080);
}

export function verifyMinimumEntries() {
    const minimumEntries = Cypress.env('minimumEntries');
    it(`Making sure that mizu shows at least ${minimumEntries} entries`, async function () {
        cy.get('#total-entries').then(number => {
            const getNum = () => {
                const numOfEntries = number.text();
                return parseInt(numOfEntries);
            };
            cy.wrap({there: getNum}).invoke('there').should('be.gte', minimumEntries);
        });
    });
}

export function leftTextCheck(entryNum, path, expectedText) {
    cy.get(`#list #entry-${entryNum} ${path}`).invoke('text').should('eq', expectedText);
}

export function leftOnHoverCheck(entryNum, path, filterName) {
    cy.get(`#list #entry-${entryNum} ${path}`).trigger('mouseover');
    cy.get(`#list #entry-${entryNum} .Queryable-Tooltip`).invoke('text').should('match', new RegExp(filterName));
}

export function rightTextCheck(path, expectedText) {
    cy.get(`#rightSideContainer ${path}`).should('have.text', expectedText);
}

export function rightOnHoverCheck(path, expectedText) {
    cy.get(`#rightSideContainer ${path}`).trigger('mouseover');
    cy.get(`#rightSideContainer .Queryable-Tooltip`).invoke('text').should('match', new RegExp(expectedText));
}

export function checkThatAllEntriesShown() {
    cy.get('#entries-length').then(number => {
        if (number.text() === '1')
            cy.get('[title="Fetch old records"]').click();
    });
}
