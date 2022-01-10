const columns = {podName : 1, namespace : 2, tapping : 3}
const greenStatusImageSrc = '/static/media/success.662997eb.svg'

export function getDomPathInStatusBar(line, column) {
    return `.expandedStatusBar > :nth-child(2) > > :nth-child(2) > :nth-child(${line}) > :nth-child(${column})`
}

export function checkLine(line, expectedValues) {
    cy.get(getDomPathInStatusBar(line, columns.podName)).invoke('text').then(podValue => {
        const podName = podValue.substring(0, podValue.indexOf('-'))
        expect(podName).to.equal(expectedValues.podName)

        cy.get(getDomPathInStatusBar(line, columns.namespace)).invoke('text').then(namespaceValue => {
            expect(namespaceValue).to.equal(expectedValues.namespace)
            cy.get(getDomPathInStatusBar(line, columns.tapping)).children().should('have.attr', 'src', greenStatusImageSrc)
        })
    })
}

export function findLineAndCheck(expectedValues) {
    cy.get('.expandedStatusBar > :nth-child(2) > > :nth-child(2) > > :nth-child(1)').then(pods => {
        cy.get('.expandedStatusBar > :nth-child(2) > > :nth-child(2) > > :nth-child(2)').then(namespaces => {

            // organizing namespaces array
            const podObjectsArray = Object.values(pods)
            const namespacesObjectsArray = Object.values(namespaces)
            let rightIndex = -1;
            namespacesObjectsArray.forEach((namespaceObj, index) => {
                rightIndex = (namespaceObj.getAttribute && namespaceObj.innerHTML === expectedValues.namespace && (podObjectsArray[index].innerHTML.substring(0, podObjectsArray[index].innerHTML.indexOf('-'))) === expectedValues.podName) ? index + 1 : rightIndex
            })
            rightIndex === -1 ? throwError(expectedValues) : checkLine(rightIndex, expectedValues);
        })
    })
}

export function throwError(expectedValues) {
    throw new Error(`The pod named ${expectedValues.podName} doesn't match any namespace named ${expectedValues.namespace}`)
}

export function getExpectedDetailsDict(podName, namespace) {
    return {podName : podName, namespace : namespace}
}
