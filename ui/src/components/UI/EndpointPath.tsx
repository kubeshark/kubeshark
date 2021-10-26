import miscStyles from "./style/misc.module.sass";
import React from "react";
import styles from './style/EndpointPath.module.sass';

interface EndpointPathProps {
    method: string
    path: string
    updateQuery: any
}

export const EndpointPath: React.FC<EndpointPathProps> = ({method, path, updateQuery}) => {
    return <div className={styles.container}>
        {method && <span
            title="Method"
            className={`${miscStyles.protocol} ${miscStyles.method}`}
            onClick={() => {
                updateQuery(`method == "${method}"`)
            }}
            style={{cursor: "pointer"}}
        >
            {method}
        </span>}
        {path && <div
            title="Summary"
            className={styles.path}
            onClick={() => {
                updateQuery(`path == "${path}"`)
            }}
            style={{cursor: "pointer"}}
        >
            {path}
        </div>}
    </div>
};
