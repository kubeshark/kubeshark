import React from 'react';
import { AutoRepresentation } from './AutoRepresentation';

interface Props {
    representation: any;
    isRulesEnabled: boolean;
    rulesMatched: any;
    color: string;
    elapsedTime: number;
}

const EntryViewer: React.FC<Props> = ({ representation, isRulesEnabled, rulesMatched, elapsedTime, color }) => {
    return <AutoRepresentation
        representation={representation}
        isRulesEnabled={isRulesEnabled}
        rulesMatched={rulesMatched}
        elapsedTime={elapsedTime}
        color={color}
        isDisplayReplay={true}
    />
};

export default EntryViewer;
