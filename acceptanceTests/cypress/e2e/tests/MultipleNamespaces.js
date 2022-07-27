import {findLineAndCheck, getExpectedDetailsDict} from '../testHelpers/StatusBarHelper';

it('opening', function () {
    cy.visit(Cypress.env('testUrl'));
    cy.get(`[data-cy="podsCountText"]`).trigger('mouseover');
});

[1, 2, 3].map(doItFunc);

function doItFunc(number) {
    const podName = Cypress.env(`name${number}`);
    const namespace = Cypress.env(`namespace${number}`);

    it(`verifying the pod (${podName}, ${namespace})`, function () {
        findLineAndCheck(getExpectedDetailsDict(podName, namespace));
    });
}
