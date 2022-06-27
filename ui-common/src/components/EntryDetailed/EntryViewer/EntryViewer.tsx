import React, { } from 'react';
import { AutoRepresentation } from './AutoRepresentation';



interface Props {
    representation: any;
    isRulesEnabled: boolean;
    rulesMatched: any;
    contractStatus: number;
    requestReason: string;
    responseReason: string;
    contractContent: string;
    color: string;
    elapsedTime: number;
}

const EntryViewer: React.FC<Props> = ({ representation, isRulesEnabled, rulesMatched, contractStatus, requestReason, responseReason, contractContent, elapsedTime, color }) => {
    return <AutoRepresentation
        representation={representation}
        isRulesEnabled={isRulesEnabled}
        rulesMatched={rulesMatched}
        contractStatus={contractStatus}
        requestReason={requestReason}
        responseReason={responseReason}
        contractContent={contractContent}
        elapsedTime={elapsedTime}
        color={color}
        isDisplayReplay={true}
    />
};

export default EntryViewer;
