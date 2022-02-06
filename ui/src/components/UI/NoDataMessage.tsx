import React from "react";
import circleImg from '../assets/dotted-circle.svg';
import './style/NoDataMessage.sass';

export interface Props {
    messageText: string;
}

const NoDataMessage: React.FC<Props> = ({messageText="No data found"}) => {

    return (
        <div className="message-container__no-data">
            <div className="container">
                <img src={circleImg} alt="No data Found"></img>
                <div className="message-container__no-data-message">{messageText}</div>
            </div>
        </div>
    );
};

export default NoDataMessage;
