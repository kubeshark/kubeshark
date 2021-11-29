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
    updateQuery: any
}

const Protocol: React.FC<ProtocolProps> = ({protocol, horizontal, updateQuery}) => {
    if (horizontal) {
        return <Queryable
            query={protocol.macro}
            updateQuery={updateQuery}
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
            updateQuery={updateQuery}
            displayIconOnMouseOver={true}
            flipped={false}
            iconStyle={{marginTop: "48px"}}
        >
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
        </Queryable>
    }
};

export default Protocol;
