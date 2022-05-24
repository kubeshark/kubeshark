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
    const entriesSent = Cypress.env('entriesCount');
    const minimumEntries = Math.round((0.75 * entriesSent));

    it(`Making sure that mizu shows at least ${minimumEntries} entries`, function () {
        cy.get('#total-entries').then(number => {
            const getNum = () => {
                return parseInt(number.text());
            };

            cy.wrap({num: getNum}).invoke('num').should('be.gt', minimumEntries);
        });
    });
}

export function leftTextCheck(entryId, path, expectedText) {
    cy.get(`#list #entry-${entryId} ${path}`).invoke('text').should('eq', expectedText);
}

export function leftOnHoverCheck(entryId, path, filterName) {
    cy.get(`#list #entry-${entryId} ${path}`).trigger('mouseover');
    cy.get(`#list #entry-${entryId} [data-cy='QueryableTooltip']`).invoke('text').should('match', new RegExp(filterName));
}

export function rightTextCheck(path, expectedText) {
    cy.get(`#rightSideContainer ${path}`).should('have.text', expectedText);
}

export function rightOnHoverCheck(path, expectedText) {
    cy.get(`#rightSideContainer ${path}`).trigger('mouseover');
    cy.get(`#rightSideContainer [data-cy='QueryableTooltip']`).invoke('text').should('match', new RegExp(expectedText));
}

export function checkFilterByMethod(funcDict) {
    const {protocol, method, methodQuery, summary, summaryQuery, numberOfRecords} = funcDict;
    const summaryDict = getSummaryDict(summary, summaryQuery);
    const methodDict = getMethodDict(method, methodQuery);
    const protocolDict = getProtocolDict(protocol.name, protocol.text);

    it(`Testing the method: ${method}`, function () {
        // applying filter
        cy.get('.w-tc-editor-text').clear().type(methodQuery);
        cy.get('[type="submit"]').click();
        cy.get('.w-tc-editor').should('have.attr', 'style').and('include', Cypress.env('greenFilterColor'));

        waitForFetch(numberOfRecords);
        pauseStream();

        cy.get(`#list [id^=entry]`).then(elements => {
            const listElmWithIdAttr = Object.values(elements);
            let doneCheckOnFirst = false;

            cy.get('#entries-length').invoke('text').then(len => {
                listElmWithIdAttr.forEach(entry => {
                    if (entry?.id && entry.id.match(RegExp(/entry-(\d{24})$/gm))) {
                        const entryId = getEntryId(entry.id);

                        leftTextCheck(entryId, methodDict.pathLeft, methodDict.expectedText);
                        leftTextCheck(entryId, protocolDict.pathLeft, protocolDict.expectedTextLeft);
                        if (summaryDict)
                            leftTextCheck(entryId, summaryDict.pathLeft, summaryDict.expectedText);

                        if (!doneCheckOnFirst) {
                            deepCheck(funcDict, protocolDict, methodDict, entry);
                            doneCheckOnFirst = true;
                        }
                    }
                });
            });
        });
    });
}

export const refreshWaitTimeout = 10000;

export function waitForFetch(gt) {
    cy.get('#entries-length', {timeout: refreshWaitTimeout}).should((el) => {
        expect(parseInt(el.text().trim(), 10)).to.be.greaterThan(gt);
    });
}

export function pauseStream() {
    cy.get('#pause-icon').click();
    cy.get('#pause-icon').should('not.be.visible');
}


export function getEntryId(id) {
    // take the second part from the string (entry-<ID>)
    return id.split('-')[1];
}

function deepCheck(generalDict, protocolDict, methodDict, entry) {
    const entryId = getEntryId(entry.id);
    const {summary, value} = generalDict;
    const summaryDict = getSummaryDict(summary);

    leftOnHoverCheck(entryId, methodDict.pathLeft, methodDict.expectedOnHover);
    leftOnHoverCheck(entryId, protocolDict.pathLeft, protocolDict.expectedOnHover);
    if (summaryDict)
        leftOnHoverCheck(entryId, summaryDict.pathLeft, summaryDict.expectedOnHover);

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
            // temporary fix, change to some "data-cy" attribute,
            // this will fix the issue that happen because we have "response:" in the header of the right side
            cy.get('#rightSideContainer > :nth-child(3)').contains('Response').click();
        cy.get(Cypress.env('bodyJsonClass')).then(text => {
            expect(text.text()).to.match(value.regex)
        });
    }
}

function getSummaryDict(value, query) {
    if (value) {
        return {
            pathLeft: '> :nth-child(2) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
            pathRight: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
            expectedText: value,
            expectedOnHover: query
        };
    }
    else {
        return null;
    }
}

function getMethodDict(value, query) {
    return {
        pathLeft: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
        pathRight: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
        expectedText: value,
        expectedOnHover: query
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
