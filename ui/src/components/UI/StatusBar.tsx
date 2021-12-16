import './style/StatusBar.sass';
import React, {useState} from "react";
import warningIcon from '../assets/warning_icon.svg';
import failIcon from '../assets/failed.svg';
import successIcon from '../assets/success.svg';

export interface TappingStatusPod {
    name: string;
    namespace: string;
    isTapped: boolean;
}

export interface TappingStatus {
    pods: TappingStatusPod[];
}

export interface Props {
    tappingStatus: TappingStatusPod[]
}

const pluralize = (noun: string, amount: number) => {
    return `${noun}${amount !== 1 ? 's' : ''}`
}

export const StatusBar: React.FC<Props> = ({tappingStatus}) => {

    const [expandedBar, setExpandedBar] = useState(false);

    const uniqueNamespaces = Array.from(new Set(tappingStatus.map(pod => pod.namespace)));
    const amountOfPods = tappingStatus.length;
    const amountOfTappedPods = tappingStatus.filter(pod => pod.isTapped).length;
    const amountOfUntappedPods = amountOfPods - amountOfTappedPods;

    return <div className={'statusBar' + (expandedBar ? ' expandedStatusBar' : "")} onMouseOver={() => setExpandedBar(true)} onMouseLeave={() => setExpandedBar(false)}>
        <div className="podsCount">
            {tappingStatus.some(pod => !pod.isTapped) && <img src={warningIcon} alt="warning"/>}
            {`Tapping ${amountOfUntappedPods > 0 ? amountOfTappedPods + " / " + amountOfPods : amountOfPods} ${pluralize('pod', amountOfPods)} in ${pluralize('namespace', uniqueNamespaces.length)} ${uniqueNamespaces.join(", ")}`}</div>
        {expandedBar && <div style={{marginTop: 20}}>
            <table>
                <thead>
                    <tr>
                        <th>Pod name</th>
                        <th>Namespace</th>
                        <th style={{marginLeft: 10}}>Tapping</th>
                    </tr>
                </thead>
                <tbody>
                    {tappingStatus.map(pod => <tr key={pod.name}>
                        <td>{pod.name}</td>
                        <td>{pod.namespace}</td>
                        <td style={{textAlign: "center"}}><img style={{height: 20}} alt="status" src={pod.isTapped ? successIcon : failIcon}/></td>
                    </tr>)}
                </tbody>
            </table>
        </div>}
    </div>;
}
