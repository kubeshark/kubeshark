export function isValueExistsInElement(shouldInclude, content, domPathToContainer){
    it(`should ${shouldInclude ? '' : 'not'} include '${content}'`, function () {
        cy.get(domPathToContainer).then(htmlText => {
            const allTextString = htmlText.text();
            if (allTextString.includes(content) !== shouldInclude)
                throw new Error(`One of the containers part contains ${content}`)
        });
    });
}

export function resizeToHugeMizu() {
    cy.viewport(1920, 3500);
}

export function resizeToNormalMizu() {
    cy.viewport(1920, 1080);
}

