import React from "react";
import styles from './style/Protocol.module.sass';
import Queryable from "./Queryable";
import variables from '../../variables.module.scss';

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
        const child = <span
            className={`${styles.base} ${styles.vertical}`}
            style={{
                backgroundColor: protocol.backgroundColor,
                color: protocol.foregroundColor,
                fontSize: protocol.fontSize,
            }}
            title={protocol.longName}
        >
            {protocol.abbr}
        </span>;

        return <Queryable
            query={protocol.macro}
            updateQuery={updateQuery}
            displayIconOnMouseOver={true}
            flipped={true}
            style={{
                backgroundColor: variables.dataBackgroundColor,
            }}
            iconStyle={{marginRight: "28px"}}
        >
            {child}
        </Queryable>
    }
};

export default Protocol;
