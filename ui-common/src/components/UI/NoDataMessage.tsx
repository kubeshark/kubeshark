import React from "react";
import circleImg from 'assets/dotted-circle.svg';
import styles from './style/NoDataMessage.module.sass'

export interface Props {
    messageText: string;
}

const NoDataMessage: React.FC<Props> = ({ messageText = "No data found" }) => {
    return (
        <div data-cy="noDataMessage" className={styles.messageContainer__noData}>
            <div className={styles.container}>
                <img src={circleImg} alt="No data Found"></img>
                <div className={styles.messageContainer__noDataMessage}>{messageText}</div>
            </div>
        </div>
    );
};

export default NoDataMessage;
