export const valueTabs = {
    response: 'RESPONSE',
    request: 'REQUEST',
    none: null
}

const maxEntriesInDom = 13;

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
    cy.viewport(Cypress.env('mizuWidth'), Cypress.env('hugeMizuHeight'));
}

export function resizeToNormalMizu() {
    cy.viewport(Cypress.env('mizuWidth'), Cypress.env('normalMizuHeight'));
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

export function checkFilterByMethod(funcDict) {
    const {protocol, method, summary, hugeMizu} = funcDict;
    const summaryDict = getSummeryDict(summary);
    const methodDict = getMethodDict(method);
    const protocolDict = getProtocolDict(protocol.name, protocol.text);

    it(`Testing the method: ${method}`, function () {
        // applying filter
        cy.get('.w-tc-editor-text').clear().type(`method == "${method}"`);
        cy.get('[type="submit"]').click();
        cy.get('.w-tc-editor').should('have.attr', 'style').and('include', Cypress.env('greenFilterColor'));

        cy.get('#entries-length').then(number => {
            // if the entries list isn't expanded it expands here
            if (number.text() === '0' || number.text() === '1') // todo change when TRA-4262 is fixed
                cy.get('[title="Fetch old records"]').click();

            cy.get('#entries-length').should('not.have.text', '0').and('not.have.text', '1').then(() => {
                cy.get(`#list [id]`).then(elements => {
                    const listElmWithIdAttr = Object.values(elements);
                    let doneCheckOnFirst = false;

                    cy.get('#entries-length').invoke('text').then(len => {
                        resizeIfNeeded(len);
                        listElmWithIdAttr.forEach(entry => {
                            if (entry?.id && entry.id.match(RegExp(/entry-(\d{2}|\d{1})$/gm))) {
                                const entryNum = getEntryNumById(entry.id);

                                leftTextCheck(entryNum, methodDict.pathLeft, methodDict.expectedText);
                                leftTextCheck(entryNum, protocolDict.pathLeft, protocolDict.expectedTextLeft);
                                if (summaryDict)
                                    leftTextCheck(entryNum, summaryDict.pathLeft, summaryDict.expectedText);

                                if (!doneCheckOnFirst) {
                                    deepCheck(funcDict, protocolDict, methodDict, entry);
                                    doneCheckOnFirst = true;
                                }
                            }
                        });
                        resizeIfNeeded(len);
                    });
                });
            });
        });
    });
}

function resizeIfNeeded(entriesLen) {
    if (entriesLen > maxEntriesInDom){
        Cypress.config().viewportHeight === Cypress.env('normalMizuHeight') ?
            resizeToHugeMizu() : resizeToNormalMizu()
    }
}

function deepCheck(generalDict, protocolDict, methodDict, entry) {
    const entryNum = getEntryNumById(entry.id);
    const {summary, value} = generalDict;
    const summaryDict = getSummeryDict(summary);

    leftOnHoverCheck(entryNum, methodDict.pathLeft, methodDict.expectedOnHover);
    leftOnHoverCheck(entryNum, protocolDict.pathLeft, protocolDict.expectedOnHover);
    if (summaryDict)
        leftOnHoverCheck(entryNum, summaryDict.pathLeft, summaryDict.expectedOnHover);

    cy.get(`#${entry.id}`).click();

    rightTextCheck(methodDict.pathRight, methodDict.expectedText);
    rightTextCheck(protocolDict.pathRight, protocolDict.expectedTextRight);
    if (summaryDict)
        rightTextCheck(summaryDict.pathRight, summaryDict.expectedText);

    rightOnHoverCheck(methodDict.pathRight, methodDict.expectedOnHover);
    rightOnHoverCheck(protocolDict.pathRight, protocolDict.expectedOnHover);
    if (summaryDict)
        rightOnHoverCheck(summaryDict.pathRight, summaryDict.expectedOnHover);

    if (value) {
        if (value.tab === valueTabs.response)
            cy.contains('Response').click();
        cy.get(Cypress.env('bodyJsonClass')).then(text => {
            expect(text.text()).to.match(value.regex)
        });
    }
}

function getSummeryDict(summary) {
    if (summary) {
        return {
            pathLeft: '> :nth-child(2) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
            pathRight: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
            expectedText: summary,
            expectedOnHover: `summary == "${summary}"`
        };
    }
    else {
        return null;
    }
}

function getMethodDict(method) {
    return {
        pathLeft: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
        pathRight: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
        expectedText: method,
        expectedOnHover: `method == "${method}"`
    };
}

function getProtocolDict(protocol, protocolText) {
    return {
        pathLeft: '> :nth-child(1) > :nth-child(1)',
        pathRight: '> :nth-child(1) > :nth-child(1) > :nth-child(1) > :nth-child(1)',
        expectedTextLeft: protocol.toUpperCase(),
        expectedTextRight: protocolText,
        expectedOnHover: protocol.toLowerCase()
    };
}

function getEntryNumById (id) {
    return parseInt(id.split('-')[1]);
}
