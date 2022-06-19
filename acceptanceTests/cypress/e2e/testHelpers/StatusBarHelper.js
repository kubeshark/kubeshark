const columns = {podName : 1, namespace : 2, tapping : 3};

function getDomPathInStatusBar(line, column) {
    return `[data-cy="expandedStatusBar"] > :nth-child(2) > > :nth-child(2) > :nth-child(${line}) > :nth-child(${column})`;
}

export function checkLine(line, expectedValues) {
    cy.get(getDomPathInStatusBar(line, columns.podName)).invoke('text').then(podValue => {
        const podName = getOnlyPodName(podValue);
        expect(podName).to.equal(expectedValues.podName);

        cy.get(getDomPathInStatusBar(line, columns.namespace)).invoke('text').then(namespaceValue => {
            expect(namespaceValue).to.equal(expectedValues.namespace);
            cy.get(getDomPathInStatusBar(line, columns.tapping)).children().should('have.attr', 'src').and("match", /success.*\.svg/);
        });
    });
}

export function findLineAndCheck(expectedValues) {
    cy.get('[data-cy="expandedStatusBar"] > :nth-child(2) > > :nth-child(2) > > :nth-child(1)').then(pods => {
        cy.get('[data-cy="expandedStatusBar"] > :nth-child(2) > > :nth-child(2) > > :nth-child(2)').then(namespaces => {
            // organizing namespaces array
            const podObjectsArray = Object.values(pods ?? {});
            const namespacesObjectsArray = Object.values(namespaces ?? {});
            let lineNumber = -1;
            namespacesObjectsArray.forEach((namespaceObj, index) => {
                const currentLine = index + 1;
                lineNumber = (namespaceObj.getAttribute && namespaceObj.innerHTML === expectedValues.namespace && (getOnlyPodName(podObjectsArray[index].innerHTML)) === expectedValues.podName) ? currentLine : lineNumber;
            });
            lineNumber === -1 ? throwError(expectedValues) : checkLine(lineNumber, expectedValues);
        });
    });
}

function throwError(expectedValues) {
    throw new Error(`The pod named ${expectedValues.podName} doesn't match any namespace named ${expectedValues.namespace}`);
}

export function getExpectedDetailsDict(podName, namespace) {
    return {podName : podName, namespace : namespace};
}

function getOnlyPodName(podElementFullStr) {
    return podElementFullStr.substring(0, podElementFullStr.indexOf('-'));
}
