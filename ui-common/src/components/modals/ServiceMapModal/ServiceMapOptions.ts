const ServiceMapOptions = {
    physics: {
        enabled: true,
        solver: 'barnesHut',
        barnesHut: {
            theta: 0.5,
            gravitationalConstant: -2000,
            centralGravity: 0.3,
            springLength: 180,
            springConstant: 0.04,
            damping: 0.09,
            avoidOverlap: 0
        },
    },
    layout: {
        hierarchical: false,
        randomSeed: 1 // always on node 1
    },
    nodes: {
        shape: 'dot',
        chosen: true,
        color: {
            background: '#27AE60',
            border: '#000000',
            highlight: {
                background: '#27AE60',
                border: '#000000',
            },
        },
        font: {
            color: '#343434',
            size: 14, // px
            face: 'arial',
            background: 'none',
            strokeWidth: 0, // px
            strokeColor: '#ffffff',
            align: 'center',
            multi: false
        },
        borderWidth: 1.5,
        borderWidthSelected: 2.5,
        labelHighlightBold: true,
        opacity: 1,
        shadow: true,
    },
    edges: {
        chosen: true,
        dashes: false,
        arrowStrikethrough: false,
        arrows: {
            to: {
                enabled: true,
            },
            middle: {
                enabled: false,
            },
            from: {
                enabled: false,
            }
        },
        smooth: {
            enabled: true,
            type: 'dynamic',
            roundness: 1.0
        },
        font: {
            color: '#343434',
            size: 12, // px
            face: 'arial',
            background: 'rgba(255,255,255,0.7)',
            strokeWidth: 2, // px
            strokeColor: '#ffffff',
            align: 'horizontal',
            multi: false
        },
        labelHighlightBold: true,
        selectionWidth: 1,
        shadow: true,
    },
    autoResize: true,
};

export default ServiceMapOptions
