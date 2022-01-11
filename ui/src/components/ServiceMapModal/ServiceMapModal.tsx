import React, { useState, useEffect, useCallback } from "react";
import { Box, Fade, Modal, Backdrop, Button } from "@material-ui/core";
import Api from "../../helpers/api";
import spinnerStyle from '../style/Spinner.module.sass';
import spinnerImg from '../assets/spinner.svg';
import Graph from "react-graph-vis";
import variables from '../../variables.module.scss';
import debounce from 'lodash/debounce';

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
}

const api = Api.getInstance();

export const ServiceMapModal: React.FC<ServiceMapModalProps> = ({ isOpen, onClose }) => {
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
        height: "760px"
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

    const getData = useCallback(async () => {
        console.log("getData called")
        try {
            setIsLoading(true)

            const serviceMapData: ServiceMapGraph = await api.serviceMapData()
            console.log(serviceMapData)

            if (serviceMapData.nodes != null) {
                for (let i = 0; i < serviceMapData.nodes.length; i++) {
                    graphData.nodes.push({
                        id: serviceMapData.nodes[i].id,
                        label: serviceMapData.nodes[i].name
                    });
                }
            }

            if (serviceMapData.edges != null) {
                for (let i = 0; i < serviceMapData.edges.length; i++) {
                    graphData.edges.push({
                        from: serviceMapData.edges[i].source.id,
                        to: serviceMapData.edges[i].destination.id
                    });
                }
            }

            setGraphData(graphData)
            setIsLoading(false)

        } catch (error) {
            setIsLoading(false)
            console.error(error);
        }
      }, [graphData])

      useEffect(() => {
        getData()
      }, [getData])

    const resetServiceMap = debounce(async () => {
        try {
            const serviceMapResetResponse = await api.serviceMapReset();
            if (serviceMapResetResponse["status"] === "enabled") {
                // close modal
                onClose()
            }

        } catch (error) {
            console.error(error);
        }
    }, 500);
    

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
                    {!isLoading && <div>
                        <Button
                            variant="contained"
                            style={{
                                margin: "0px 0px 0px 0px",
                                backgroundColor: variables.blueColor,
                                fontWeight: 600,
                                borderRadius: "4px",
                                color: "#fff",
                                textTransform: "none",
                            }}
                            onClick={resetServiceMap}
                        >
                            Reset
                        </Button>
                        <Graph
                            graph={graphData}
                            options={options}
                        />
                    </div>}
                </Box>
            </Fade>
        </Modal>
    );

}