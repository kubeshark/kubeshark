import React from "react";
import styles from './style/Protocol.module.sass';

export interface ProtocolInterface {
    name: string
    longName: string
    abbr: string
    backgroundColor: string
    foregroundColor: string
    fontSize: number
    referenceLink: string
    ports: string[]
    inbound_ports: string
}

interface ProtocolProps {
    protocol: ProtocolInterface
    horizontal: boolean
}

const Protocol: React.FC<ProtocolProps> = ({protocol, horizontal}) => {
    if (horizontal) {
        return <a target="_blank" rel="noopener noreferrer" href={protocol.referenceLink}>
            <span
                className={`${styles.base} ${styles.horizontal}`}
                style={{
                    backgroundColor: protocol.backgroundColor,
                    color: protocol.foregroundColor,
                    fontSize: 13,
                }}
                title={protocol.abbr}
            >
                {protocol.longName}
            </span>
        </a>
    } else {
        return <a target="_blank" rel="noopener noreferrer" href={protocol.referenceLink}>
            <span
                className={`${styles.base} ${styles.vertical}`}
                style={{
                    backgroundColor: protocol.backgroundColor,
                    color: protocol.foregroundColor,
                    fontSize: protocol.fontSize,
                }}
                title={protocol.longName}
            >
                {protocol.abbr}
            </span>
        </a>
    }
};

export default Protocol;
