import React, { useState, useEffect } from "react";
import { Box, Fade, Modal, Backdrop } from "@material-ui/core";
import Api from "../../helpers/api";
import spinnerStyle from '../style/Spinner.module.sass';
import spinnerImg from '../assets/spinner.svg';
import ForceGraph2D, { GraphData } from 'react-force-graph-2d';


interface ServiceMapModalProps {
    isOpen: boolean;
    onClose: () => void;
    api: Api
}

interface ServiceMapNode {
    name: string;
    protocol: string;
    count: number;
}

interface ServiceMapEdge {
    source: string;
    destination: string;
    count: number;
}

interface ServiceMapGraph {
    nodes: ServiceMapNode[];
    edges: ServiceMapEdge[];
}

export const ServiceMapModal: React.FC<ServiceMapModalProps> = ({ isOpen, onClose, api }) => {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [graphData, setGraphData] = useState<GraphData>({
        nodes: [],
        links: []
    });

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
                let data: GraphData = {
                    nodes: [],
                    links: []
                }

                for (let i = 0; i < serviceMapData.nodes.length; i++) {
                    data.nodes.push({
                        id: serviceMapData.nodes[i].name
                    });
                }

                for (let i = 0; i < serviceMapData.edges.length; i++) {
                    data.links.push({
                        source: serviceMapData.edges[i].source,
                        target: serviceMapData.edges[i].destination
                    });
                }


                setGraphData(data)
                setIsLoading(false)

            } catch (error) {
                console.error(error);
            }
        })()
    }, [api]);

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
                {/* <Box sx={style}> */}
                <div>
                    {isLoading && <div className={spinnerStyle.spinnerContainer}>
                        <img alt="spinner" src={spinnerImg} style={{ height: 50 }} />
                    </div>}
                    {!isLoading && <ForceGraph2D
                        graphData={graphData}
                        linkDirectionalArrowLength={3.5}
                        linkDirectionalArrowRelPos={1}
                        linkCurvature={0.25}
                    />}
                </div>
                {/* </Box> */}
            </Fade>
        </Modal>
    );

}