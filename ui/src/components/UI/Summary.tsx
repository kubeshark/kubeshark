import miscStyles from "./style/misc.module.sass";
import React from "react";
import styles from './style/Summary.module.sass';
import Queryable from "./Queryable";

interface SummaryProps {
    method: string
    summary: string
}

export const Summary: React.FC<SummaryProps> = ({method, summary}) => {

    return <div className={styles.container}>
        {method && <Queryable
            query={`method == "${method}"`}
            className={`${miscStyles.protocol} ${miscStyles.method}`}
            displayIconOnMouseOver={true}
            style={{whiteSpace: "nowrap"}}
            flipped={true}
            iconStyle={{zIndex:"5",position:"relative",right:"22px"}}
        >
            <span>
                {method}
            </span>
        </Queryable>}
        {summary && <Queryable
            query={`summary == "${summary}"`}
            displayIconOnMouseOver={true}
            flipped={true}
            iconStyle={{zIndex:"5",position:"relative",right:"14px"}}
        >
            <div
                className={`${styles.summary}`}
            >
                {summary}
            </div>
        </Queryable>}
    </div>
};
