export function check(shouldInclude, content, domPathToContainer){
    it(`should ${shouldInclude ? '' : 'not'} include '${content}'`, function () {
        cy.get(domPathToContainer).then(body => {
            const allTextString = body.text();
            if (allTextString.includes(content) !== shouldInclude)
                throw new Error(`One of the containers part contains ${content}`)
        });
    });
}
