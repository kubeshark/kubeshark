it('should open mizu', function () {
    cy.visit(Cypress.env('testUrl'));
});

serviceMapCheck();

async function serviceMapCheck() {
    it('service map test', function () {
        cy.intercept(`${Cypress.env('testUrl')}/servicemap/get`).as('serviceMapRequest');
        cy.get('#total-entries').should('not.have.text', '0').then(() => {
            cy.get('#total-entries').invoke('text').then(async entriesNum => {
                cy.get('[alt="service-map"]').click();
                cy.wait('@serviceMapRequest').then(({response}) => {
                    const body = response.body;
                    console.log(body)
                    // checkBody(body, {
                    //     nodesCount: 2,
                    //     mainNodeName: 'redis.mizu-tests',
                    //     entiresNum: entriesNum,
                    //     edgesCount: 1,
                    //     protocol: 'redis'
                    // });
                    redisTestServiceMapValidation(body, parseInt(entriesNum));
                });
            });
        });
    });
}


function redisTestServiceMapValidation(bodyJson, entriesNum) {
    const {nodes, edges, status} = bodyJson;
    // const {nodesCount, mainNodeName, entriesNum, edgesCount, protocol} = valuesDict;
    const nodesNum = 2;
    const protocolNodeName = 'redis.mizu-tests';

    expect(nodes.length).to.equal(nodesNum, `Expected nodes count`);

    const {protocolNode, protocolNodeIndex} = nodes[0].name === protocolNodeName ?
        {protocolNode: nodes[0].name, index: 0} : {protocolNode: nodes[1].name, protocolNodeIndex: 1};

    expect(edges[0].destination.name).to.equal(protocolNodeName);
    expect(edges[0].source.name).to.equal(protocolNodeIndex ? nodes[0].name : nodes[1].name);

    let falseEntriesCount = entriesNum;
    edges.forEach(edge => {
        falseEntriesCount -= edge.count;
    });

    if (falseEntriesCount) {
        throw new Error(``)
    }

    // let tempNodesCount = 0;
    // nodes.forEach(node => {
    //     if (node.name === mainNodeName){
    //         tempNodesCount++;
    //     }
    // });
    // if (tempNodesCount !== 1) {
    //     throw new Error(`Expected only one rabbit nodes. got: ${tempNodesCount}`);
    // }
    //
    // expect(edges.length).to.equal(edgesCount, `Expected edges count`);
    // let tempNum = 0;
    // edges.forEach(edge => {
    //     expect(edge.protocol.name).to.equal(protocol);
    //     tempNum += edge.count;
    // });
    // if (tempNum !== 30) {
    //     throw new Error(`Expected the counts sum to be ${entriesNum}, but got ${tempNum} instead`);
    // }



}
