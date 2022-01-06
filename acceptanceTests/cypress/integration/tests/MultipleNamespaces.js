const columns = {"podName" : 1, "namespace" : 2, "tapping" : 3}
const greenStatusImageSrc = "/static/media/success.662997eb.svg"
it('verifying the first pod', function() {
    cy.visit('http://localhost:8899/')
    cy.get('.podsCount').trigger('mouseover')
    findLineAndCheck({"podName" : Cypress.env('name1'), "namespace" : Cypress.env('namespace1')})
})

it('verifying the second pod', function () {
    findLineAndCheck({"podName" : Cypress.env('name2'), "namespace" : Cypress.env('namespace2')})
})

function getDomPathInStatusBar(line, column) {
    return `.expandedStatusBar > :nth-child(2) > > :nth-child(2) > :nth-child(${line}) > :nth-child(${column})`
}

function checkLine(line, expectedValues) {

    cy.get(getDomPathInStatusBar(line, columns.podName)).invoke('text').then(podValue => {
        const podName = podValue.substring(0, podValue.indexOf('-'))
        expect(podName).to.equal(expectedValues.podName)

        cy.get(getDomPathInStatusBar(line, columns.namespace)).invoke('text').then(namespaceValue => {
            expect(namespaceValue).to.equal(expectedValues.namespace)

            cy.get(getDomPathInStatusBar(line, columns.tapping)).children().should('have.attr', 'src', greenStatusImageSrc)
        })
    })
}

function findLineAndCheck(expectedValues) {
    cy.get('.expandedStatusBar > :nth-child(2) > > :nth-child(2) > > :nth-child(1)').then(pods => {
        cy.get('.expandedStatusBar > :nth-child(2) > > :nth-child(2) > > :nth-child(2)').then(namespaces => {

            // organizing namespaces array
            const namespacesObjectsArray = Object.values(namespaces)
            let namespacesArray = []
            namespacesObjectsArray.forEach(line => {
                line.getAttribute ? namespacesArray.push(line.innerHTML) : null
            })

            // organizing pods array
            const podObjectsArray = Object.values(pods)
            let podsArray = []
            podObjectsArray.forEach(line => {
                line.getAttribute ? podsArray.push(line.innerHTML.substring(0, line.innerHTML.indexOf('-'))) : null
            })

            let rightIndex = -1
            podsArray.forEach((element, index) => {
                if (element === expectedValues.podName && namespacesArray[index] === expectedValues.namespace) {
                    rightIndex = index + 1
                }
            })
            rightIndex === -1 ? throwError(expectedValues.podName, expectedValues.namespace) : checkLine(rightIndex, expectedValues)
        })
    })
}
function throwError(pod, namespace) {
    throw new Error(`The pod named ${pod} doesn't match any namespace named ${namespace}`)
}
