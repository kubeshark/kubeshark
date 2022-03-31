import React from "react";
import styles from './style/Protocol.module.sass';
import Queryable from "./Queryable";

export interface ProtocolInterface {
    name: string
    longName: string
    abbr: string
    macro: string
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
        return <Queryable
            query={protocol.macro}
            displayIconOnMouseOver={true}
        >
            <a target="_blank" rel="noopener noreferrer" href={protocol.referenceLink}>
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
        </Queryable>
    } else {
        return <Queryable
            query={protocol.macro}
            displayIconOnMouseOver={true}
            flipped={false}
            iconStyle={{marginTop: "52px", marginRight: "10px", zIndex: 1000}}
            tooltipStyle={{marginTop: "-22px", zIndex: 1001}}
        >
            <span
                className={`${styles.base} ${styles.vertical}`}
                style={{
                    backgroundColor: protocol.backgroundColor,
                    color: protocol.foregroundColor,
                    fontSize: protocol.fontSize,
                    marginRight: "-6px",
                }}
                title={protocol.longName}
            >
                {protocol.abbr}
            </span>
        </Queryable>
    }
};

export default Protocol;
