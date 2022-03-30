
const minNodeScaling = 10
const maxNodeScaling = 30

const minEdgeScaling = 1
const maxEdgeScaling = maxNodeScaling / 2

const minLabelScaling = 11
const maxLabelScaling = 16
const selectedNodeColor = "#0C0B1A"
const selectedNodeBorderColor = "#205CF5"
const selectedNodeLabelColor = "#205CF5"
const selectedEdgeLabelColor = "#205CF5"

const customScaling = (min, max, total, value) => {
    if (max === min) {
        return 0.5;
    }
    else {
        const scale = 1 / (max - min);
        return Math.max(0, (value - min) * scale);
    }
}

const nodeSelected = (values, id, selected, hovering) => {
    values.color = selectedNodeColor;
    values.borderColor = selectedNodeBorderColor;
    values.borderWidth = 4;
}

const nodeLabelSelected = (values, id, selected, hovering) => {
    values.size = values.size + 1;
    values.color = selectedNodeLabelColor;
    values.strokeColor = selectedNodeLabelColor;
    values.strokeWidth = 0.2
}

const edgeSelected = (values, id, selected, hovering) => {
    values.opacity = 0.4;
    values.width = values.width + 1;
}

const edgeLabelSelected = (values, id, selected, hovering) => {
    values.size = values.size + 1;
    values.color = selectedEdgeLabelColor;
    values.strokeColor = selectedEdgeLabelColor;
    values.strokeWidth = 0.2
}

const nodeOptions = {
    shape: 'dot',
    chosen: {
        node: nodeSelected,
        label: nodeLabelSelected,
    },
    color: {
        background: '#494677',
        border: selectedNodeColor,
    },
    font: {
        color: selectedNodeColor,
        size: 11, // px
        face: 'Roboto',
        background: '#FFFFFFBF',
        strokeWidth: 0.2, // px
        strokeColor: selectedNodeColor,
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
        customScalingFunction: customScaling,
    },
    borderWidth: 2,
    labelHighlightBold: true,
    opacity: 1,
    shadow: true,
}

const edgeOptions = {
    chosen: {
        edge: edgeSelected,
        label: edgeLabelSelected,
    },
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
        customScalingFunction: customScaling,
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
            centralGravity: 0.4,
            springLength: 180,
            springConstant: 0.04,
            damping: 0.2,
            avoidOverlap: 0
        },
    },
    layout: {
        hierarchical: false,
        randomSeed: 1 // always on node 1
    },
    nodes: nodeOptions,
    edges: edgeOptions,
    autoResize: true,
    interaction: {
        selectable: true,
        selectConnectedEdges: true,
        multiselect: true,
        dragNodes: true,
        dragView: true,
        hover: true,
        hoverConnectedEdges: true,
        zoomView: true,
        zoomSpeed: 1,
    },
};

export default ServiceMapOptions
