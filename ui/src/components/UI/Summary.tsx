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
            text={method}
            query={`method == "${method}"`}
            updateQuery={updateQuery}
            title="Method"
            className={`${miscStyles.protocol} ${miscStyles.method}`}
            wrapperStyle={{height: "14px"}}
            applyTextEllipsis={false}
            displayIconOnMouseOver={true}
        />}
        {summary && <Queryable
            text={summary}
            query={`summary == "${summary}"`}
            updateQuery={updateQuery}
            title="Summary"
            className={`${styles.summary}`}
            applyTextEllipsis={false}
            displayIconOnMouseOver={true}
        />}
    </div>
};
