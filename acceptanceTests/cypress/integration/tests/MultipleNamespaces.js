import {findLineAndCheck} from '../page_objects/StatusBar';
import {getExpectedDetailsDict} from '../page_objects/StatusBar';

it('opening', function () {
    cy.visit(Cypress.env('testUrl'));
    cy.get('.podsCount').trigger('mouseover');
});

[1, 2, 3].map(doItFunc);

function doItFunc(number) {
    const podName = Cypress.env(`name${number}`);
    const namespace = Cypress.env(`namespace${number}`);

    it(`verifying the pod (${podName}, ${namespace})`, function () {
        findLineAndCheck(getExpectedDetailsDict(podName, namespace))
    })
}

