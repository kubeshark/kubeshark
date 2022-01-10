it('Loading mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

['redact', 'REDACT', '{ "User": "[REDACTED]" }', '{"User":"[REDACTED]"}'].map(doItFunc)

function doItFunc(stringToCheck) {
    it(`No '${stringToCheck}' should exist`, function () {
        cy.get('.CollapsibleContainer', { timeout : 10 * 1000 }).then(containersText => {
            const allText = containersText.text();
            if (allText.includes(stringToCheck))
                throw new Error(`The containers include the string: ${stringToCheck}`);
        });
    });
}

