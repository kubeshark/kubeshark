
const minNodeScaling = 10
const maxNodeScaling = 30
const minEdgeScaling = 1
const maxEdgeScaling = maxNodeScaling / 2
const minLabelScaling = 11
const maxLabelScaling = 16

const nodeOptions = {
    shape: 'dot',
    chosen: true,
    color: {
        background: '#494677',
        border: '#0C0B1A',
        highlight: {
            background: '#0C0B1A',
            border: '#205CF5',
        },
    },
    font: {
        color: '#0C0B1A',
        size: 11, // px
        face: 'Roboto',
        background: '#FFFFFFBF',
        strokeWidth: 0.2, // px
        strokeColor: '#0C0B1A',
        align: 'center',
        multi: false,
    },
    // defines the node min and max sizes when zoom in/out, based on the node value
    scaling: {
        min: minNodeScaling,
        max: maxNodeScaling,
        // defines the label scaling size in px
        label: {
            enabled: true,
            min: minLabelScaling,
            max: maxLabelScaling,
            maxVisible: maxLabelScaling,
            drawThreshold: 5,
        },
        customScalingFunction: function (min, max, total, value) {
            if (max === min) {
                return 0.5;
            }
            else {
                let scale = 1 / (max - min);
                return Math.max(0, (value - min) * scale);
            }
        }
    },
    borderWidth: 2,
    borderWidthSelected: 4,
    labelHighlightBold: true,
    opacity: 1,
    shadow: true,
}

const edgeOptions = {
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
        color: '#000000',
        size: 11, // px
        face: 'Roboto',
        background: '#FFFFFFCC',
        strokeWidth: 0.2, // px
        strokeColor: '#000000',
        align: 'horizontal',
        multi: false,
    },
    scaling: {
        min: minEdgeScaling,
        max: maxEdgeScaling,
        label: {
            enabled: true,
            min: minLabelScaling,
            max: maxLabelScaling,
            maxVisible: maxLabelScaling,
            drawThreshold: 5
        },
        customScalingFunction: function (min, max, total, value) {
            if (max === min) {
                return 0.5;
            }
            else {
                var scale = 1 / (max - min);
                return Math.max(0, (value - min) * scale);
            }
        }
    },
    labelHighlightBold: true,
    selectionWidth: 1,
    shadow: true,
}

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
            avoidOverlap: 1
        },
    },
    layout: {
        hierarchical: false,
        randomSeed: 1 // always on node 1
    },
    nodes: nodeOptions,
    edges: edgeOptions,
    autoResize: true,
};

export default ServiceMapOptions