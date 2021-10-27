import miscStyles from "./style/misc.module.sass";
import React from "react";
import styles from './style/Summary.module.sass';

interface SummaryProps {
    method: string
    summary: string
    updateQuery: any
}

export const Summary: React.FC<SummaryProps> = ({method, summary, updateQuery}) => {
    return <div className={styles.container}>
        {method && <span
            title="Method"
            className={`queryable ${miscStyles.protocol} ${miscStyles.method}`}
            onClick={() => {
                updateQuery(`method == "${method}"`)
            }}
        >
            {method}
        </span>}
        {summary && <div
            title="Summary"
            className={`queryable ${styles.summary}`}
            onClick={() => {
                updateQuery(`summary == "${summary}"`)
            }}
        >
            {summary}
        </div>}
    </div>
};
