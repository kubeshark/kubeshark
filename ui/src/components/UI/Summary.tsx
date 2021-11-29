import miscStyles from "./style/misc.module.sass";
import React from "react";
import styles from './style/Summary.module.sass';
import Queryable from "./Queryable";

interface SummaryProps {
    method: string
    summary: string
    updateQuery: any
}

export const Summary: React.FC<SummaryProps> = ({method, summary, updateQuery}) => {

    return <div className={styles.container}>
        {method && <Queryable
            query={`method == "${method}"`}
            className={`${miscStyles.protocol} ${miscStyles.method}`}
            updateQuery={updateQuery}
            displayIconOnMouseOver={true}
        >
            <span>
                {method}
            </span>
        </Queryable>}
        {summary && <Queryable
            query={`summary == "${summary}"`}
            updateQuery={updateQuery}
            displayIconOnMouseOver={true}
        >
            <div
                className={`${styles.summary}`}
            >
                {summary}
            </div>
        </Queryable>}
    </div>
};
