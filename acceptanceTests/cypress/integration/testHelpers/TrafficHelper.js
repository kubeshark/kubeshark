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
    cy.get(`#list #entry-${entryNum} .Queryable-Tooltip`).should('have.text', filterName);
}

export function rightTextCheck(path, expectedText) {
    cy.get(`.TrafficPage-Container > :nth-child(2) ${path}`).should('have.text', expectedText);
}

export function rightOnHoverCheck(path, expectedText) {
    cy.get(`.TrafficPage-Container > :nth-child(2) ${path}`).trigger('mouseover');
    cy.get(`.TrafficPage-Container > :nth-child(2) .Queryable-Tooltip`).should('have.text', expectedText);
}
