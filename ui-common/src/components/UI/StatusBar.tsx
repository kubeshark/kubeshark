import style from './style/StatusBar.module.sass';
import React, {useState} from "react";
import warningIcon from 'assets/warning_icon.svg';
import failIcon from 'assets/failed.svg';
import successIcon from 'assets/success.svg';
import {useRecoilValue} from "recoil";
import tappingStatusAtom, {tappingStatusDetails} from "../../recoil/tappingStatus";

const pluralize = (noun: string, amount: number) => {
    return `${noun}${amount !== 1 ? 's' : ''}`
}

export const StatusBar = () => {

    const tappingStatus = useRecoilValue(tappingStatusAtom);
    const [expandedBar, setExpandedBar] = useState(false);
    const {uniqueNamespaces, amountOfPods, amountOfTappedPods, amountOfUntappedPods} = useRecoilValue(tappingStatusDetails);

    return <div className={`${style.statusBar} ${(expandedBar ? ' expandedStatusBar' : "")}`} onMouseOver={() => setExpandedBar(true)} onMouseLeave={() => setExpandedBar(false)}>
        <div className={style.podsCount}>
            {tappingStatus?.some(pod => !pod.isTapped) && <img src={warningIcon} alt="warning"/>}
            {`Tapping ${amountOfUntappedPods > 0 ? amountOfTappedPods + " / " + amountOfPods : amountOfPods} ${pluralize('pod', amountOfPods)} in ${pluralize('namespace', uniqueNamespaces.length)} ${uniqueNamespaces.join(", ")}`}
            </div>
        {expandedBar && <div style={{marginTop: 20}}>
            <table>
                <thead>
                    <tr>
                        <th style={{width: "40%"}}>Pod name</th>
                        <th style={{width: "40%"}}>Namespace</th>
                        <th style={{width: "20%", textAlign: "center"}}>Tapping</th>
                    </tr>
                </thead>
                <tbody>
                    {tappingStatus.map(pod => <tr key={pod.name}>
                        <td style={{width: "40%"}}>{pod.name}</td>
                        <td style={{width: "40%"}}>{pod.namespace}</td>
                        <td style={{width: "20%", textAlign: "center"}}><img style={{height: 20}} alt="status" src={pod.isTapped ? successIcon : failIcon}/></td>
                    </tr>)}
                </tbody>
            </table>
        </div>}
    </div>;
}
