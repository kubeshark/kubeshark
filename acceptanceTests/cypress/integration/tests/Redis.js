import {
    checkThatAllEntriesShown,
    leftOnHoverCheck,
    leftTextCheck,
    rightOnHoverCheck,
    rightTextCheck,
    verifyMinimumEntries
} from "../testHelpers/TrafficHelper";

it('opening mizu', function () {
    cy.visit(Cypress.env('testUrl'));

});

['PING', 'SET', 'EXISTS', 'GET', 'DEL'].map(checkRedisFilterByMethod);

function checkRedisFilterByMethod(method) {
    it(`Testing the method: ${method}`, function () {
        // applying filter
        cy.get('.w-tc-editor-text').clear().type(`method == "${method}"`);
        cy.get('[type="submit"]').click();
        cy.get('.w-tc-editor').should('have.attr', 'style').and('include', Cypress.env('greenFilterColor'));

        cy.get('#entries-length').then(number => {
            // if the entries list isn't expanded it expands here
            if (number.text() === '0' || number.text() === '1')
                cy.get('[title="Fetch old records"]').click();

            cy.get('#entries-length').should('not.have.text', '0').and('not.have.text', '1').then(() => { //TODO remove the 0 when bug fixed
                cy.get('#list [id]').then(elements => {
                    const htmlMnt = Object.values(elements);
                    let doneOneClickCheck = false;

                    htmlMnt.forEach((elm, index) => {
                        // going through every element in the list that has attribute id that equals: "entry-X"
                        if (elm?.id && elm.id.match(RegExp(/entry-(\d{2}|\d{1})$/gm))) {
                            // dictionary for saving the status of each check
                            let containsTheRightAttr = {
                                checkByTitle: false,
                                checkByMethod: false
                            };

                            // going through every element inside the entry
                            elm.querySelectorAll("*").forEach(child => {
                                if (child.getAttribute('title') ===
                                    `Add "method == "${method}"" to the filter`) {
                                    containsTheRightAttr.checkByTitle = true;
                                }
                                if (child.innerText === method)
                                    containsTheRightAttr.checkByMethod = true;
                            });
                            if (!containsTheRightAttr.checkByTitle || !containsTheRightAttr.checkByMethod){
                                throw new Error(`Failed. ${containsTheRightAttr}`);
                            }
                            if (!doneOneClickCheck && method !== 'PING') {
                                cy.get(`#list #${elm.id}`).click();
                                const entryNum = parseInt(elm.id.split('-')[1]);
                                const pathToSummary = '> :nth-child(2) > :nth-child(1) > :nth-child(2) > :nth-child(2)';
                                const redisElement = '[title="REDIS"]';

                                leftOnHoverCheck(entryNum + 1, pathToSummary, 'summary == "key"');
                                leftTextCheck(entryNum + 1, pathToSummary, 'key');

                                rightOnHoverCheck(redisElement, 'redis');
                                rightTextCheck(redisElement, 'Redis Serialization Protocol');

                                if (method !== 'SET') {
                                    cy.contains('Response').click();
                                }

                                rightValueCheck(method);
                                doneOneClickCheck = true;
                            }
                        }
                    });
                    cy.reload();
                });
            });
        });
    });
}
function rightValueCheck(method) {
    let valueToCheck;

    switch (method) {
        case 'SET':
            valueToCheck = /^\[value, keepttl]$/mg;
            break;
        case 'EXISTS':
            valueToCheck = /^1$/mg;
            break;
        case 'GET':
            valueToCheck = /^value$/mg;
            break;
        case 'DEL':
            valueToCheck = /^1$|^0$/mg;
            break;
    }
    cy.get('.hljs').then(text => {
        expect(text.text()).to.match(valueToCheck);
    });
}

function rightSideSETCheck() {
    cy.get('.hljs').should('have.text', '[value, keepttl]');
}

function rightSideEXISTSCheck() {
    //cy.contains('Response').click();
    cy.get('.hljs').should('have.text', '1');
}

function rightSideGETCheck() {
    //cy.contains('Response').click();
    cy.get('.hljs').should('have.text', 'value');
}

function rightSideDELCheck() {
    //cy.contains('Response').click();
    cy.get('.hljs').should('have.text', '1')
}

