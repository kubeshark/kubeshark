import React, { useState, useEffect } from "react";
import { Box, Fade, Modal, Backdrop } from "@material-ui/core";
import Api from "../../helpers/api";
import spinnerStyle from '../style/Spinner.module.sass';
import spinnerImg from '../assets/spinner.svg';
import Graph from "react-graph-vis";


interface GraphData {
    nodes: Node[];
    edges: Edge[];
}

interface Node {
    id: number;
    label: string;
    title?: string;
    color?: string;
}

interface Edge {
    from: number;
    to: number;
}

interface ServiceMapNode {
    name: string;
    id: number;
    protocol: string;
    count: number;
}

interface ServiceMapEdge {
    source: ServiceMapNode;
    destination: ServiceMapNode;
    count: number;
}

interface ServiceMapGraph {
    nodes: ServiceMapNode[];
    edges: ServiceMapEdge[];
}

interface ServiceMapModalProps {
    isOpen: boolean;
    onClose: () => void;
    api: Api
}

export const ServiceMapModal: React.FC<ServiceMapModalProps> = ({ isOpen, onClose, api }) => {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [graphData, setGraphData] = useState<GraphData>({
        nodes: [],
        edges: []
    });

    const options = {
        layout: {
          hierarchical: true
        },
        edges: {
          color: "#000000"
        },
        height: "500px"
      };

    const style = {
        position: 'absolute',
        top: '10%',
        left: '50%',
        transform: 'translate(-50%, 0%)',
        width: '80vw',
        bgcolor: 'background.paper',
        borderRadius: '5px',
        boxShadow: 24,
        p: 4,
        color: '#000',
    };

    useEffect(() => {
        (async () => {
            try {
                const serviceMapData: ServiceMapGraph = await api.serviceMapData()

                for (let i = 0; i < serviceMapData.nodes.length; i++) {
                    graphData.nodes.push({
                        id: serviceMapData.nodes[i].id,
                        label: serviceMapData.nodes[i].name
                    });
                }

                for (let i = 0; i < serviceMapData.edges.length; i++) {
                    graphData.edges.push({
                        from: serviceMapData.edges[i].source.id,
                        to: serviceMapData.edges[i].destination.id
                    });
                }

                setGraphData(graphData)
                setIsLoading(false)

            } catch (error) {
                console.error(error);
            }
        })()
    }, [api, graphData]);

    return (
        <Modal
            aria-labelledby="transition-modal-title"
            aria-describedby="transition-modal-description"
            open={isOpen}
            onClose={onClose}
            closeAfterTransition
            BackdropComponent={Backdrop}
            BackdropProps={{
                timeout: 500,
            }}
            style={{ overflow: 'auto' }}
        >
            <Fade in={isOpen}>
                <Box sx={style}>
                    {isLoading && <div className={spinnerStyle.spinnerContainer}>
                        <img alt="spinner" src={spinnerImg} style={{ height: 50 }} />
                    </div>}
                    {!isLoading && <Graph
                        graph={graphData}
                        options={options}
                    />}
                </Box>
            </Fade>
        </Modal>
    );

}