import {findLineAndCheck, getExpectedDetailsDict} from "../testHelpers/StatusBarHelper";
import {
    leftTextCheck,
    resizeToHugeMizu,
    resizeToNormalMizu,
    rightOnHoverCheck,
    rightTextCheck, verifyMinimumEntries
} from "../testHelpers/TrafficHelper";
const refreshWaitTimeout = 10000;
const bodyJsonClass = '.hljs';

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

verifyMinimumEntries();

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

it('right side sanity test', function () {
    cy.get('#entryDetailedTitleBodySize').then(sizeTopLine => {
        const sizeOnTopLine = sizeTopLine.text().replace(' B', '');
        cy.contains('Response').click();
        cy.contains('Body Size (bytes)').parent().next().then(size => {
            const bodySizeByDetails = size.text();
            expect(sizeOnTopLine).to.equal(bodySizeByDetails, 'The body size in the top line should match the details in the response');

            if (parseInt(bodySizeByDetails) < 0) {
                throw new Error(`The body size cannot be negative. got the size: ${bodySizeByDetails}`)
            }

            cy.get('#entryDetailedTitleElapsedTime').then(timeInMs => {
                const time = timeInMs.text();
                if (time < '0ms') {
                    throw new Error(`The time in the top line cannot be negative ${time}`);
                }

                cy.get('#rightSideContainer [title="Status Code"]').then(status => {
                    const statusCode = status.text();
                    cy.contains('Status').parent().next().then(statusInDetails => {
                        const statusCodeInDetails = statusInDetails.text();

                        expect(statusCode).to.equal(statusCodeInDetails, 'The status code in the top line should match the status code in details');
                    });
                });
            });
        });
    });
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
            cy.get('.w-tc-editor').should('have.attr', 'style').and('include', Cypress.env('greenFilterColor'));
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
            cy.get('.w-tc-editor').should('have.attr', 'style').and('include', Cypress.env('redFilterColor'));
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
            cy.get('.w-tc-editor').should('have.attr', 'style').and('include', Cypress.env('greenFilterColor'));
            if (!applyByEnter)
                cy.get('[type="submit"]').click();

            // only one entry in DOM after filtering, checking all checks on it
            leftTextCheck(totalEntries - 1, leftSidePath, leftSideExpectedText);
            leftOnHoverCheck(totalEntries - 1, leftSidePath, name);
            rightTextCheck(rightSidePath, rightSideExpectedText);
            rightOnHoverCheck(rightSidePath, name);
            checkRightSideResponseBody();

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

function checkRightSideResponseBody() {
    cy.contains('Response').click();
    clickCheckbox('Decode Base64');

    cy.get(`${bodyJsonClass}`).then(value => {
        const encodedBody = value.text();
        const decodedBody = atob(encodedBody);
        const responseBody = JSON.parse(decodedBody);

        const expectdJsonBody = {
            args: RegExp({}),
            url: RegExp('http://.*/get'),
            headers: {
                "User-Agent": RegExp('[REDACTED]'),
                "Accept-Encoding": RegExp('gzip'),
                "X-Forwarded-Uri": RegExp('/api/v1/namespaces/.*/services/.*/proxy/get')
            }
        };

        expect(responseBody.args).to.match(expectdJsonBody.args);
        expect(responseBody.url).to.match(expectdJsonBody.url);
        expect(responseBody.headers['User-Agent']).to.match(expectdJsonBody.headers['User-Agent']);
        expect(responseBody.headers['Accept-Encoding']).to.match(expectdJsonBody.headers['Accept-Encoding']);
        expect(responseBody.headers['X-Forwarded-Uri']).to.match(expectdJsonBody.headers['X-Forwarded-Uri']);

        cy.get(`${bodyJsonClass}`).should('have.text', encodedBody);
        clickCheckbox('Decode Base64');

        cy.get(`${bodyJsonClass} > `).its('length').should('be.gt', 1).then(linesNum => {
            cy.get(`${bodyJsonClass} > >`).its('length').should('be.gt', linesNum).then(jsonItemsNum => {
                checkPrettyAndLineNums(jsonItemsNum, decodedBody);

                clickCheckbox('Line numbers');
                checkPrettyOrNothing(jsonItemsNum, decodedBody);

                clickCheckbox('Pretty');
                checkPrettyOrNothing(jsonItemsNum, decodedBody);

                clickCheckbox('Line numbers');
                checkOnlyLineNumberes(jsonItemsNum, decodedBody);
            });
        });
    });
}

function clickCheckbox(type) {
    cy.contains(`${type}`).prev().children().click();
}

function checkPrettyAndLineNums(jsonItemsLen, decodedBody) {
    decodedBody = decodedBody.replaceAll(' ', '');
    cy.get(`${bodyJsonClass} >`).then(elements => {
        const lines = Object.values(elements);
        lines.forEach((line, index) => {
            if (line.getAttribute) {
                const cleanLine = getCleanLine(line);
                const currentLineFromDecodedText = decodedBody.substring(0, cleanLine.length);

                expect(cleanLine).to.equal(currentLineFromDecodedText, `expected the text in line number ${index + 1} to match the text that generated by the base64 decoding`)

                decodedBody = decodedBody.substring(cleanLine.length);
            }
        });
    });
}

function getCleanLine(lineElement) {
    return (lineElement.innerText.substring(0, lineElement.innerText.length - 1)).replaceAll(' ', '');
}

function checkPrettyOrNothing(jsonItems, decodedBody) {
    cy.get(`${bodyJsonClass} > `).should('have.length', jsonItems).then(text => {
        const json = text.text();
        expect(json).to.equal(decodedBody);
    });
}

function checkOnlyLineNumberes(jsonItems, decodedText) {
    cy.get(`${bodyJsonClass} >`).should('have.length', 1).and('have.text', decodedText);
    cy.get(`${bodyJsonClass} > >`).should('have.length', jsonItems)
}
