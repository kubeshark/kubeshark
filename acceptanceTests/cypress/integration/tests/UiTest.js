import {findLineAndCheck, getExpectedDetailsDict} from "../testHelpers/StatusBarHelper";
import {resizeToHugeMizu, resizeToNormalMizu} from "../testHelpers/TrafficHelper";
const greenFilterColor = 'rgb(210, 250, 210)';
const redFilterColor = 'rgb(250, 214, 220)';
const refreshWaitTimeout = 10000;

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

it('top bar check', function () {
    const podName1 = 'httpbin', namespace1 = 'mizu-tests';
    const podName2 = 'httpbin2', namespace2 = 'mizu-tests';

    cy.get('.podsCount').trigger('mouseover');
    findLineAndCheck(getExpectedDetailsDict(podName1, namespace1));
    findLineAndCheck(getExpectedDetailsDict(podName2, namespace2));
    cy.reload();
});

it('filtering guide check', function () {
    cy.get('[title="Open Filtering Guide (Cheatsheet)"]').click();
    cy.get('#modal-modal-title').should('be.visible');
    cy.get('[lang="en"]').click(0, 0);
    cy.get('#modal-modal-title').should('not.exist');
});

checkIllegalFilter('invalid filter');

checkFilter({
    name: 'http',
    leftSidePath: '> :nth-child(1) > :nth-child(1)',
    leftSideExpectedText: 'HTTP',
    rightSidePath: '[title=HTTP]',
    rightSideExpectedText: 'Hypertext Transfer Protocol -- HTTP/1.1',
    applyByEnter: true
});

checkFilter({
    name: 'response.status == 200',
    leftSidePath: '[title="Status Code"]',
    leftSideExpectedText: '200',
    rightSidePath: '> :nth-child(2) [title="Status Code"]',
    rightSideExpectedText: '200',
    applyByEnter: false
});

checkFilter({
    name: 'src.name == ""',
    leftSidePath: '[title="Source Name"]',
    leftSideExpectedText: '[Unresolved]',
    rightSidePath: '> :nth-child(2) [title="Source Name"]',
    rightSideExpectedText: '[Unresolved]',
    applyByEnter: false
});

checkFilter({
    name: 'method == "GET"',
    leftSidePath: '> :nth-child(3) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
    leftSideExpectedText: 'GET',
    rightSidePath: '> :nth-child(2) > :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
    rightSideExpectedText: 'GET',
    applyByEnter: true
});

checkFilter({
    name: 'summary == "/get"',
    leftSidePath: '> :nth-child(3) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
    leftSideExpectedText: '/get',
    rightSidePath: '> :nth-child(2) > :nth-child(2) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
    rightSideExpectedText: '/get',
    applyByEnter: false
});

checkFilter({
    name: 'dst.name == "httpbin.mizu-tests"',
    leftSidePath: '> :nth-child(3) > :nth-child(2) > :nth-child(3) > :nth-child(2)',
    leftSideExpectedText: 'httpbin.mizu-tests',
    rightSidePath: '> :nth-child(2) > :nth-child(2) > :nth-child(2) > :nth-child(3) > :nth-child(2)',
    rightSideExpectedText: 'httpbin.mizu-tests',
    applyByEnter: false
});

checkFilter({
    name: 'src.ip == "127.0.0.1"',
    leftSidePath: '[title="Source IP"]',
    leftSideExpectedText: '127.0.0.1',
    rightSidePath: '> :nth-child(2) [title="Source IP"]',
    rightSideExpectedText: '127.0.0.1',
    applyByEnter: false
});

checkFilterNoResults('method == "POST"');

function checkFilterNoResults(filterName) {
    it(`checking the filter: ${filterName}. Expecting no results`, function () {
        cy.get('#total-entries').then(number => {
            const totalEntries = number.text();

            // applying the filter
            cy.get('.w-tc-editor-text').type(filterName);
            cy.get('.w-tc-editor').should('have.attr', 'style').and('include', greenFilterColor);
            cy.get('[type="submit"]').click();

            // waiting for the entries number to load
            cy.get('#total-entries', {timeout: refreshWaitTimeout}).should('have.text', totalEntries);

            // the DOM should show 0 entries
            cy.get('#entries-length').should('have.text', '0');

            // going through every potential entry and verifies that it doesn't exist
            [...Array(parseInt(totalEntries)).keys()].map(shouldNotExist);

            cy.get('[title="Fetch old records"]').click();
            cy.get('#noMoreDataTop', {timeout: refreshWaitTimeout}).should('be.visible');
            cy.get('#entries-length').should('have.text', '0'); // after loading all entries there should still be 0 entries

            // reloading then waiting for the entries number to load
            cy.reload();
            cy.get('#total-entries', {timeout: refreshWaitTimeout}).should('have.text', totalEntries);
        });
    });
}

function shouldNotExist(entryNum) {
    cy.get(`entry-${entryNum}`).should('not.exist');
}

function checkIllegalFilter(illegalFilterName) {
    it(`should show red search bar with the input: ${illegalFilterName}`, function () {
        cy.get('#total-entries').then(number => {
            const totalEntries = number.text();

            cy.get('.w-tc-editor-text').type(illegalFilterName);
            cy.get('.w-tc-editor').should('have.attr', 'style').and('include', redFilterColor);
            cy.get('[type="submit"]').click();

            cy.get('[role="alert"]').should('be.visible');
            cy.get('.w-tc-editor-text').clear();

            // reloading then waiting for the entries number to load
            cy.reload();
            cy.get('#total-entries', {timeout: refreshWaitTimeout}).should('have.text', totalEntries);
        });
    });
}
function checkFilter(filterDetails){
    const {name, leftSidePath, rightSidePath, rightSideExpectedText, leftSideExpectedText, applyByEnter} = filterDetails;
    const entriesForDeeperCheck = 5;

    it(`checking the filter: ${name}`, function () {
        cy.get('#total-entries').then(number => {
            const totalEntries = number.text();

            // checks the hover on the last entry (the only one in DOM at the beginning)
            leftOnHoverCheck(totalEntries - 1, leftSidePath, name);

            // applying the filter with alt+enter or with the button
            cy.get('.w-tc-editor-text').type(`${name}${applyByEnter ? '{alt+enter}' : ''}`);
            cy.get('.w-tc-editor').should('have.attr', 'style').and('include', greenFilterColor);
            if (!applyByEnter)
                cy.get('[type="submit"]').click();

            // only one entry in DOM after filtering, checking all four checks on it
            leftTextCheck(totalEntries - 1, leftSidePath, leftSideExpectedText);
            leftOnHoverCheck(totalEntries - 1, leftSidePath, name);
            rightTextCheck(rightSidePath, rightSideExpectedText);
            rightOnHoverCheck(rightSidePath, name);

            cy.get('[title="Fetch old records"]').click();
            resizeToHugeMizu();

            // waiting for the entries number to load
            cy.get('#entries-length', {timeout: refreshWaitTimeout}).should('have.text', totalEntries);

            // checking only 'leftTextCheck' on all entries because the rest of the checks require more time
            [...Array(parseInt(totalEntries)).keys()].forEach(entryNum => {
                leftTextCheck(entryNum, leftSidePath, leftSideExpectedText);
            });

            // making the other 3 checks on the first X entries (longer time for each check)
            deeperChcek(leftSidePath, rightSidePath, name, leftSideExpectedText, rightSideExpectedText, entriesForDeeperCheck);

            // reloading then waiting for the entries number to load
            resizeToNormalMizu();
            cy.reload();
            cy.get('#total-entries', {timeout: refreshWaitTimeout}).should('have.text', totalEntries);
        });
    });
}

function deeperChcek(leftSidePath, rightSidePath, filterName, leftSideExpectedText, rightSideExpectedText, entriesNumToCheck) {
    [...Array(entriesNumToCheck).keys()].forEach(entryNum => {
        leftOnHoverCheck(entryNum, leftSidePath, filterName);

        cy.get(`#list #entry-${entryNum}`).click();
        rightTextCheck(rightSidePath, rightSideExpectedText);
        rightOnHoverCheck(rightSidePath, filterName);
    });
}

function leftTextCheck(entryNum, path, expectedText) {
    cy.get(`#list #entry-${entryNum} ${path}`).invoke('text').should('eq', expectedText);
}

function leftOnHoverCheck(entryNum, path, filterName) {
    cy.get(`#list #entry-${entryNum} ${path}`).trigger('mouseover');
    cy.get(`#list #entry-${entryNum} .Queryable-Tooltip`).should('have.text', filterName);
}

function rightTextCheck(path, expectedText) {
    cy.get(`.TrafficPage-Container > :nth-child(2) ${path}`).should('have.text', expectedText);
}

function rightOnHoverCheck(path, expectedText) {
    cy.get(`.TrafficPage-Container > :nth-child(2) ${path}`).trigger('mouseover');
    cy.get(`.TrafficPage-Container > :nth-child(2) .Queryable-Tooltip`).should('have.text', expectedText);
}
