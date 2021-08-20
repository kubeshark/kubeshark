import React from "react";
import internal from "stream";
import styles from './style/Protocol.module.sass';

export interface ProtocolInterface {
    name: string,
    long_name: string,
    abbreviation: string,
    background_color: string,
    foreground_color: string,
    font_size: number,
    reference_link: string,
    outbound_ports: string[],
    inbound_ports: string,
}

interface ProtocolProps {
    protocol: ProtocolInterface
}

const Protocol: React.FC<ProtocolProps> = ({protocol}) => {
    return <a target="_blank" rel="noopener noreferrer" href={protocol.reference_link}>
        <span
            className={`${styles.base}`}
            style={{
                backgroundColor: protocol.background_color,
                color: protocol.foreground_color,
                fontSize: protocol.font_size,
            }}
            title={protocol.long_name}
        >
            {protocol.abbreviation}
        </span>
    </a>
};

export default Protocol;
