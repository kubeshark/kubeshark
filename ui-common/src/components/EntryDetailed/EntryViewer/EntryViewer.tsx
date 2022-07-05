import React from 'react';
import { AutoRepresentation } from './AutoRepresentation';

interface Props {
    representation: any;
    color: string;
}

const EntryViewer: React.FC<Props> = ({representation, color}) => {
    return <AutoRepresentation
        representation={representation}
        color={color}
        isDisplayReplay={true}
    />
};

export default EntryViewer;
