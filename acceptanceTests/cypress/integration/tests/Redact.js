const expectedLineJson = '{ "User": "[REDACTED]" }'

it('should', function () {
    cy.visit(Cypress.env('testUrl'));
})

it('checking in Headers', function () {
    cy.get('.CollapsibleContainer', { timeout : 15 * 1000}).first().next().then(elements => {
        const allText = elements.text();
        if (!allText.includes('User-Header[REDACTED]'))
            throw new Error('The headers panel doesnt include User-Header [REDACTED]');
    });
});

it('checking in Body', function () {
    cy.get('.hljs').then($element => {
        const line = $element.text();
        expect(line).to.equal(expectedLineJson);
    });
});
