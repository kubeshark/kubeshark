import {
    leftOnHoverCheck,
    leftTextCheck,
    rightOnHoverCheck,
    rightTextCheck,
} from "../testHelpers/TrafficHelper";

const valueTabs = {
    response: 'RESPONSE',
    request: 'REQUEST',
    none: null
}

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

checkRedisFilterByMethod({
    method: 'PING',
    shouldCheckSummary: false,
    valueTab: valueTabs.none
});

checkRedisFilterByMethod({
    method: 'SET',
    shouldCheckSummary: true,
    valueTab: valueTabs.request,
    valueRegex: /^\[value, keepttl]$/mg
});

checkRedisFilterByMethod({
    method: 'EXISTS',
    shouldCheckSummary: true,
    valueTab: valueTabs.response,
    valueRegex: /^1$/mg
});

checkRedisFilterByMethod({
    method: 'GET',
    shouldCheckSummary: true,
    valueTab: valueTabs.response,
    valueRegex: /^value$/mg
});

checkRedisFilterByMethod({
    method: 'DEL',
    shouldCheckSummary: true,
    valueTab: valueTabs.response,
    valueRegex: /^1$|^0$/mg
});

function checkRedisFilterByMethod(funcDict) {
    const {method, shouldCheckSummary} = funcDict
    const summaryDict = getSummeryDict();
    const methodDict = getMethodDict(method);
    const protocolDict = getProtocolDict();

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

                   listElmWithIdAttr.forEach(entry => {
                       if (entry?.id && entry.id.match(RegExp(/entry-(\d{2}|\d{1})$/gm))) {
                           const entryNum = getEntryNumById(entry.id);

                           leftTextCheck(entryNum, methodDict.pathLeft, methodDict.expectedText);
                           leftTextCheck(entryNum, protocolDict.pathLeft, protocolDict.expectedTextLeft);
                           if (shouldCheckSummary)
                               leftTextCheck(entryNum, summaryDict.pathLeft, summaryDict.expectedText);

                           if (!doneCheckOnFirst) {
                               deepCheck(funcDict, protocolDict, methodDict, summaryDict, entry);
                               doneCheckOnFirst = true;
                           }
                       }
                   });
                });
            });
        });
    });
}

function deepCheck(generalDict, protocolDict, methodDict, summaryDict, entry) {
    const entryNum = getEntryNumById(entry.id);
    const {shouldCheckSummary, valueTab, valueRegex} = generalDict;

    leftOnHoverCheck(entryNum, methodDict.pathLeft, methodDict.expectedOnHover);
    leftOnHoverCheck(entryNum, protocolDict.pathLeft, protocolDict.expectedOnHover);
    if (shouldCheckSummary)
        leftOnHoverCheck(entryNum, summaryDict.pathLeft, summaryDict.expectedOnHover);

    cy.get(`#${entry.id}`).click();

    rightTextCheck(methodDict.pathRight, methodDict.expectedText);
    rightTextCheck(protocolDict.pathRight, protocolDict.expectedTextRight);
    if (shouldCheckSummary)
        rightTextCheck(summaryDict.pathRight, summaryDict.expectedText);

    rightOnHoverCheck(methodDict.pathRight, methodDict.expectedOnHover);
    rightOnHoverCheck(protocolDict.pathRight, protocolDict.expectedOnHover);
    if (shouldCheckSummary)
        rightOnHoverCheck(summaryDict.pathRight, summaryDict.expectedOnHover);

    if (valueTab) {
        if (valueTab === valueTabs.response)
            cy.contains('Response').click();
        cy.get(Cypress.env('bodyJsonClass')).then(text => {
            expect(text.text()).to.match(valueRegex)
        });
    }
}

function getSummeryDict() {
    return {
        pathLeft: '> :nth-child(2) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
        pathRight: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
        expectedText: 'key',
        expectedOnHover: `summary == "key"`
    };
}

function getMethodDict(method) {
    return {
        pathLeft: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
        pathRight: '> :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
        expectedText: method,
        expectedOnHover: `method == "${method}"`
    };
}

function getProtocolDict() {
    return {
        pathLeft: '> :nth-child(1) > :nth-child(1)',
        pathRight: '> :nth-child(1) > :nth-child(1) > :nth-child(1) > :nth-child(1)',
        expectedTextLeft: 'REDIS',
        expectedTextRight: 'Redis Serialization Protocol',
        expectedOnHover: `redis`
    };
}

function getEntryNumById (id) {
    return parseInt(id.split('-')[1]);
}
