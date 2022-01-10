it('No redact test', function () {
    cy.visit(Cypress.env('testUrl'));

    cy.get('.CollapsibleContainer', { timeout : 15 * 1000} ).then(allContainersText => {
        const allText = allContainersText.text();

        ['redacted', 'REDACTED', '{ "User": "[REDACTED]" }', '{"User":"[REDACTED]"}'].map(shouldNotInclude);

        function shouldNotInclude(checkStr) {
            allText.includes(checkStr) ? expect(0).to.e

        }
    });
});


