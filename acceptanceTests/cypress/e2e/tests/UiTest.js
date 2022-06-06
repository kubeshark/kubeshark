import {findLineAndCheck, getExpectedDetailsDict} from "../testHelpers/StatusBarHelper";
import {
    getEntryId,
    leftOnHoverCheck,
    leftTextCheck,
    resizeToHugeMizu,
    resizeToNormalMizu,
    rightOnHoverCheck,
    rightTextCheck,
    verifyMinimumEntries,
    refreshWaitTimeout,
    waitForFetch,
    pauseStream
} from "../testHelpers/TrafficHelper";

const fullParam = Cypress.env('arrayDict'); // "Name:fooNamespace:barName:foo1Namespace:bar1"
const podsArray = fullParam.split('Name:').slice(1); // ["fooNamespace:bar", "foo1Namespace:bar1"]
podsArray.forEach((podStr, index) => {
    const podAndNamespaceArr = podStr.split('Namespace:'); // [foo, bar] / [foo1, bar1]
    podsArray[index] = getExpectedDetailsDict(podAndNamespaceArr[0], podAndNamespaceArr[1]);
});

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

verifyMinimumEntries();

it('top bar check', function () {
    cy.get(`[data-cy="podsCountText"]`).trigger('mouseover');
    podsArray.map(findLineAndCheck);
    cy.reload();
});

it('filtering guide check', function () {
    cy.reload();
    cy.get('[title="Open Filtering Guide (Cheatsheet)"]').click();
    cy.get('#modal-modal-title').should('be.visible');
    cy.get('[lang="en"]').click(0, 0);
    cy.get('#modal-modal-title').should('not.exist');
});

it('right side sanity test', function () {
    cy.get('#entryDetailedTitleElapsedTime').then(timeInMs => {
        const time = timeInMs.text();
        if (time < '0ms') {
            throw new Error(`The time in the top line cannot be negative ${time}`);
        }
    });

    // temporary fix, change to some "data-cy" attribute,
    // this will fix the issue that happen because we have "response:" in the header of the right side
    cy.get('#rightSideContainer > :nth-child(3)').contains('Response').click();

    cy.get('#rightSideContainer [title="Status Code"]').then(status => {
        const statusCode = status.text();
        cy.contains('Status').parent().next().then(statusInDetails => {
            const statusCodeInDetails = statusInDetails.text();

            expect(statusCode).to.equal(statusCodeInDetails, 'The status code in the top line should match the status code in details');
        });
    });
});

checkIllegalFilter('invalid filter');

checkFilter({
    filter: 'http',
    leftSidePath: '> :nth-child(1) > :nth-child(1)',
    leftSideExpectedText: 'HTTP',
    rightSidePath: '[title=HTTP]',
    rightSideExpectedText: 'Hypertext Transfer Protocol -- HTTP/1.1',
    applyByCtrlEnter: true,
    numberOfRecords: 20,
});

checkFilter({
    filter: 'response.status == 200',
    leftSidePath: '[title="Status Code"]',
    leftSideExpectedText: '200',
    rightSidePath: '> :nth-child(2) [title="Status Code"]',
    rightSideExpectedText: '200',
    applyByCtrlEnter: false,
    numberOfRecords: 20
});

if (Cypress.env('shouldCheckSrcAndDest')) {
    serviceMapCheck();

    checkFilter({
        filter: 'src.name == ""',
        leftSidePath: '[title="Source Name"]',
        leftSideExpectedText: '[Unresolved]',
        rightSidePath: '> :nth-child(2) [title="Source Name"]',
        rightSideExpectedText: '[Unresolved]',
        applyByCtrlEnter: false,
        numberOfRecords: 20
    });

    checkFilter({
        filter: `dst.name == "httpbin.mizu-tests"`,
        leftSidePath: '> :nth-child(3) > :nth-child(2) > :nth-child(3) > :nth-child(2)',
        leftSideExpectedText: 'httpbin.mizu-tests',
        rightSidePath: '> :nth-child(2) > :nth-child(2) > :nth-child(2) > :nth-child(3) > :nth-child(2)',
        rightSideExpectedText: 'httpbin.mizu-tests',
        applyByCtrlEnter: false,
        numberOfRecords: 20
    });
}

checkFilter({
    filter: 'request.method == "GET"',
    leftSidePath: '> :nth-child(3) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
    leftSideExpectedText: 'GET',
    rightSidePath: '> :nth-child(2) > :nth-child(2) > :nth-child(1) > :nth-child(1) > :nth-child(2)',
    rightSideExpectedText: 'GET',
    applyByCtrlEnter: true,
    numberOfRecords: 20
});

checkFilter({
    filter: 'request.path == "/get"',
    leftSidePath: '> :nth-child(3) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
    leftSideExpectedText: '/get',
    rightSidePath: '> :nth-child(2) > :nth-child(2) > :nth-child(1) > :nth-child(2) > :nth-child(2)',
    rightSideExpectedText: '/get',
    applyByCtrlEnter: false,
    numberOfRecords: 20
});

checkFilter({
    filter: 'src.ip == "127.0.0.1"',
    leftSidePath: '[title="Source IP"]',
    leftSideExpectedText: '127.0.0.1',
    rightSidePath: '> :nth-child(2) [title="Source IP"]',
    rightSideExpectedText: '127.0.0.1',
    applyByCtrlEnter: false,
    numberOfRecords: 20
});

checkFilterNoResults('request.method == "POST"');

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

            cy.get('[title="Fetch old records"]').click();
            cy.get('#noMoreDataTop', {timeout: refreshWaitTimeout}).should('be.visible');
            cy.get('#entries-length').should('have.text', '0'); // after loading all entries there should still be 0 entries

            // reloading then waiting for the entries number to load
            cy.reload();
            cy.get('#total-entries', {timeout: refreshWaitTimeout}).should('have.text', totalEntries);
        });
    });
}

function checkIllegalFilter(illegalFilterName) {
    it(`should show red search bar with the input: ${illegalFilterName}`, function () {
        cy.reload();
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

function checkFilter(filterDetails) {
    const {
        filter,
        leftSidePath,
        rightSidePath,
        rightSideExpectedText,
        leftSideExpectedText,
        applyByCtrlEnter,
        numberOfRecords
    } = filterDetails;

    const entriesForDeeperCheck = 5;

    it(`checking the filter: ${filter}`, function () {
        cy.get('.w-tc-editor-text').clear();
        // applying the filter with alt+enter or with the button
        cy.get('.w-tc-editor-text').type(`${filter}${applyByCtrlEnter ? '{ctrl+enter}' : ''}`);
        cy.get('.w-tc-editor').should('have.attr', 'style').and('include', Cypress.env('greenFilterColor'));
        if (!applyByCtrlEnter)
            cy.get('[type="submit"]').click();

        waitForFetch(numberOfRecords);
        pauseStream();

        cy.get(`#list [id^=entry]`).last().then(elem => {
            const element = elem[0];
            const entryId = getEntryId(element.id);

            // only one entry in DOM after filtering, checking all checks on it
            leftTextCheck(entryId, leftSidePath, leftSideExpectedText);
            leftOnHoverCheck(entryId, leftSidePath, filter);

            rightTextCheck(rightSidePath, rightSideExpectedText);
            rightOnHoverCheck(rightSidePath, filter);
            checkRightSideResponseBody();
        });

        resizeToHugeMizu();

        // checking only 'leftTextCheck' on all entries because the rest of the checks require more time
        cy.get(`#list [id^=entry]`).each(elem => {
            const element = elem[0];
            let entryId = getEntryId(element.id);
            leftTextCheck(entryId, leftSidePath, leftSideExpectedText);
        });

        // making the other 3 checks on the first X entries (longer time for each check)
        deeperCheck(leftSidePath, rightSidePath, filter, rightSideExpectedText, entriesForDeeperCheck);

        // reloading then waiting for the entries number to load
        resizeToNormalMizu();
        cy.reload();
        waitForFetch(numberOfRecords);
        pauseStream();
    });
}

function deeperCheck(leftSidePath, rightSidePath, filterName, rightSideExpectedText, entriesNumToCheck) {
    cy.get(`#list [id^=entry]`).each((element, index) => {
        if (index < entriesNumToCheck) {
            const entryId = getEntryId(element[0].id);
            leftOnHoverCheck(entryId, leftSidePath, filterName);

            cy.get(`#list #entry-${entryId}`).click();
            rightTextCheck(rightSidePath, rightSideExpectedText);
            rightOnHoverCheck(rightSidePath, filterName);
        }
    });
}

function checkRightSideResponseBody() {
    // temporary fix, change to some "data-cy" attribute,
    // this will fix the issue that happen because we have "response:" in the header of the right side
    cy.get('#rightSideContainer > :nth-child(3)').contains('Response').click();
    clickCheckbox('Decode Base64');

    cy.get(`${Cypress.env('bodyJsonClass')}`).then(value => {
        const encodedBody = value.text();
        const decodedBody = atob(encodedBody);
        const responseBody = JSON.parse(decodedBody);


        const expectdJsonBody = {
            args: RegExp({}),
            url: RegExp('http://.*/get'),
            headers: {
                "User-Agent": RegExp('client'),
                "Accept-Encoding": RegExp('gzip'),
                "X-Forwarded-Uri": RegExp('/api/v1/namespaces/.*/services/.*/proxy/get')
            }
        };

        expect(responseBody.args).to.match(expectdJsonBody.args);
        expect(responseBody.url).to.match(expectdJsonBody.url);
        expect(responseBody.headers['User-Agent']).to.match(expectdJsonBody.headers['User-Agent']);
        expect(responseBody.headers['Accept-Encoding']).to.match(expectdJsonBody.headers['Accept-Encoding']);
        expect(responseBody.headers['X-Forwarded-Uri']).to.match(expectdJsonBody.headers['X-Forwarded-Uri']);

        cy.get(`${Cypress.env('bodyJsonClass')}`).should('have.text', encodedBody);
        clickCheckbox('Decode Base64');

        cy.get(`${Cypress.env('bodyJsonClass')} > `).its('length').should('be.gt', 1).then(linesNum => {
            cy.get(`${Cypress.env('bodyJsonClass')} > >`).its('length').should('be.gt', linesNum).then(jsonItemsNum => {
                // checkPrettyAndLineNums(decodedBody);

                //clickCheckbox('Line numbers');
                //checkPrettyOrNothing(jsonItemsNum, decodedBody);

                // clickCheckbox('Pretty');
                // checkPrettyOrNothing(jsonItemsNum, decodedBody);
                //
                // clickCheckbox('Line numbers');
                // checkOnlyLineNumberes(jsonItemsNum, decodedBody);
            });
        });
    });
}

function clickCheckbox(type) {
    cy.contains(`${type}`).prev().children().click();
}

function checkPrettyAndLineNums(decodedBody) {
    decodedBody = decodedBody.replaceAll(' ', '');
    cy.get(`${Cypress.env('bodyJsonClass')} >`).then(elements => {
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
    cy.get(`${Cypress.env('bodyJsonClass')} > `).should('have.length', jsonItems).then(text => {
        const json = text.text();
        expect(json).to.equal(decodedBody);
    });
}

function checkOnlyLineNumberes(jsonItems, decodedText) {
    cy.get(`${Cypress.env('bodyJsonClass')} >`).should('have.length', 1).and('have.text', decodedText);
    cy.get(`${Cypress.env('bodyJsonClass')} > >`).should('have.length', jsonItems)
}

function serviceMapCheck() {
    it('service map test', function () {
        cy.intercept(`${Cypress.env('testUrl')}/servicemap/get`).as('serviceMapRequest');
        cy.get('#total-entries').should('not.have.text', '0').then(() => {
            cy.get('#total-entries').invoke('text').then(entriesNum => {
                cy.get('[alt="service-map"]').click();
                cy.wait('@serviceMapRequest').then(({response}) => {
                    const body = response.body;
                    const nodeParams = {
                        destination: 'httpbin.mizu-tests',
                        source: '127.0.0.1'
                    };
                    serviceMapAPICheck(body, parseInt(entriesNum), nodeParams);
                    cy.reload();
                });
            });
        });
    });
}

function serviceMapAPICheck(body, entriesNum, nodeParams) {
    const {nodes, edges} = body;

    expect(nodes.length).to.equal(Object.keys(nodeParams).length, `Expected nodes count`);

    expect(edges.some(edge => edge.source.name === nodeParams.source)).to.be.true;
    expect(edges.some(edge => edge.destination.name === nodeParams.destination)).to.be.true;

    let count = 0;
    edges.forEach(edge => {
        count += edge.count;
        if (edge.destination.name === nodeParams.destination) {
            expect(edge.source.name).to.equal(nodeParams.source);
        }
    });

    expect(count).to.equal(entriesNum);
}
