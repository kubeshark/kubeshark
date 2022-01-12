const inHeader = 'User-Header[REDACTED]';
const inBody = '{ "User": "[REDACTED]" }';
const shouldExist = Cypress.env('shouldExist');

it('Loading Mizu', function () {
    cy.visit(Cypress.env('testUrl'));
})

it(`should ${shouldExist ? '' : 'not'} include ${inHeader}`, function () {
    cy.get('.CollapsibleContainer', { timeout : 15 * 1000}).first().next().then(headerElements => {
        const allText = headerElements.text();
        if (allText.includes(inHeader) !== shouldExist)
            throw new Error(`The headers panel doesnt include ${inHeader}`);
    });
});

it(`should ${shouldExist ? '' : 'not'} include ${inBody}`, function () {
    cy.get('.hljs').then(bodyElement => {
        const line = bodyElement.text();
        if (line.includes(inBody) !== shouldExist)
            throw new Error(`The body panel doesnt include ${inBody}`);
    });
});
