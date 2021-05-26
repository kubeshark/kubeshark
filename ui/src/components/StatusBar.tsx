import './style/StatusBar.sass';
import React from "react";

export interface TappingStatusPod {
    name: string;
    namespace: string;
}

export interface TappingStatus {
    pods: TappingStatusPod[];
}

export interface Props {
    tappingStatus: TappingStatus
}

const pluralize = (noun: string, amount: number) => {
    return `${noun}${amount != 1 ? 's' : ''}`
}

export const StatusBar: React.FC<Props> = ({tappingStatus}) => {
    const uniqueNamespaces = Array.from(new Set(tappingStatus.pods.map(pod => pod.namespace)));
    const amountOfPods = tappingStatus.pods.length;

    return <div className='StatusBar'>
        <span>{`Tapping ${amountOfPods} ${pluralize('pod', amountOfPods)} in ${pluralize('namespace', uniqueNamespaces.length)} ${uniqueNamespaces.join(", ")}`}</span>
    </div>;
}
