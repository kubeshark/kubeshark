import './style/StatusBar.sass';
import React, {useState} from "react";

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
    return `${noun}${amount !== 1 ? 's' : ''}`
}

export const StatusBar: React.FC<Props> = ({tappingStatus}) => {

    const [expandedBar, setExpandedBar] = useState(false);

    const uniqueNamespaces = Array.from(new Set(tappingStatus.pods.map(pod => pod.namespace)));
    const amountOfPods = tappingStatus.pods.length;

    return <div className={'statusBar' + (expandedBar ? ' expandedStatusBar' : "")} onMouseOver={() => setExpandedBar(true)} onMouseLeave={() => setExpandedBar(false)}>
        <div className="podsCount">{`Tapping ${amountOfPods} ${pluralize('pod', amountOfPods)} in ${pluralize('namespace', uniqueNamespaces.length)} ${uniqueNamespaces.join(", ")}`}</div>
        {expandedBar && <div style={{marginTop: 20}}>
            <table>
                <thead>
                    <tr>
                        <th>Pod name</th>
                        <th>Namespace</th>
                    </tr>
                </thead>
                <tbody>
                    {tappingStatus.pods.map(pod => <tr key={pod.name}>
                        <td>{pod.name}</td>
                        <td>{pod.namespace}</td>
                    </tr>)}
                </tbody>
            </table>
        </div>}
    </div>;
}
