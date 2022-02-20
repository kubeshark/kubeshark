it('should open mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

serviceMapCheck();

function serviceMapCheck() {
    it('service map test', function () {
        cy.intercept(`${Cypress.env('testUrl')}/servicemap/get`).as('serviceMapRequest');
        cy.get('#total-entries').invoke('text').then(entriesNum => {
            cy.get('[alt="service-map"]').click();
            cy.wait('@serviceMapRequest').then(({response}) => {
                const body = response.body;
                console.log(body)
                checkBody(body, entriesNum);
            });
        });
    });
}


function checkBody(bodyJson, entriesNum) {
    const {nodes, edges, status} = bodyJson;
    expect(nodes.length).to.equal(3, `Expected nodes count`);

    let rabbitNodesCount = 0;
    nodes.forEach(node => {
        if (node.name === 'rabbitmq.mizu-tests'){
            rabbitNodesCount++;
        }
    });
    if (rabbitNodesCount !== 1) {
        throw new Error(`Expected only one rabbit nodes. got: ${rabbitNodesCount}`);
    }

    expect(edges.length).to.equal(3, `Expected edges count`);
    let tempNum = 0;
    edges.forEach(edge => {
        tempNum += edge.count;

        if (edge.destination.name === 'rabbitmq.mizu-tests') {
            expect(edge.count).to.equal(27);
        } else {
            expect(edge.source.name).to.equal('rabbitmq.mizu-tests');
        }
        expect(edge.protocol.name).to.equal('amqp');

    });
    if (tempNum !== 72) {
        throw new Error(`Expected the counts sum to be ${entriesNum}, but got ${tempNum} instead`);
    }



}
