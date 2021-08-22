import React from "react";
import styles from './style/StatusCode.module.sass';

export enum StatusCodeClassification {
    SUCCESS = "success",
    FAILURE = "failure",
    NEUTRAL = "neutral"
}

interface HAREntryProps {
    statusCode: number
}

const StatusCode: React.FC<HAREntryProps> = ({statusCode}) => {

    const classification = getClassification(statusCode)

    return <span
        title="Status Code"
        className={`${styles[classification]} ${styles.base}`}>
            {statusCode}
    </span>
};

export function getClassification(statusCode: number): string {
    let classification = StatusCodeClassification.NEUTRAL;

    if ((statusCode >= 200 && statusCode <= 399) || statusCode === 0) {
        classification = StatusCodeClassification.SUCCESS;
    } else if (statusCode >= 400 || (statusCode >= 1 && statusCode <= 16)) {
        classification = StatusCodeClassification.FAILURE;
    }

    return classification
}

export default StatusCode;
