import style from './style/StatusBar.module.sass';
import React, {useState} from "react";
import warningIcon from 'assets/warning_icon.svg';
import failIcon from 'assets/failed.svg';
import successIcon from 'assets/success.svg';
import {useRecoilValue} from "recoil";
import tappingStatusAtom, {tappingStatusDetails} from "../../recoil/tappingStatus";
import Tooltip from "./Tooltip";

const pluralize = (noun: string, amount: number) => {
    return `${noun}${amount !== 1 ? 's' : ''}`
}

interface StatusBarProps {
   isDemoBannerView: boolean;
   disabled?: boolean;
  }

export const StatusBar: React.FC<StatusBarProps> = ({isDemoBannerView, disabled}) => {
    const tappingStatus = useRecoilValue(tappingStatusAtom);
    const [expandedBar, setExpandedBar] = useState(false);
    const {uniqueNamespaces, amountOfPods, amountOfTappedPods, amountOfUntappedPods} = useRecoilValue(tappingStatusDetails);
    return <div style={{opacity: disabled ? 0.4 : 1}} className={`${isDemoBannerView ? `${style.banner}` : ''} ${style.statusBar} ${(expandedBar && !disabled ? `${style.expandedStatusBar}` : "")}`} onMouseOver={() => setExpandedBar(true)} onMouseLeave={() => setExpandedBar(false)} data-cy="expandedStatusBar">
        <div className={style.podsCount}>
          {!disabled && tappingStatus.some(pod => !pod.isTapped) && <img src={warningIcon} alt="warning"/>}
          {disabled && <Tooltip title={"Tapping status is not updated when streaming is paused"} isSimple><img src={warningIcon} alt="warning"/></Tooltip>}
            <span className={style.podsCountText} data-cy="podsCountText">
                {`Tapping ${amountOfUntappedPods > 0 ? amountOfTappedPods + " / " + amountOfPods : amountOfPods} ${pluralize('pod', amountOfPods)} in ${pluralize('namespace', uniqueNamespaces.length)} ${uniqueNamespaces.join(", ")}`}
            </span>
        </div>
        {expandedBar && !disabled && <div style={{marginTop: 20}}>
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
                        <td style={{width: "20%", textAlign: "center"}}>{pod.isTapped ? <img style={{height: 20}} alt="status" src={successIcon}/> : <img style={{height: 20}} alt="status" src={failIcon}/>}</td>
                    </tr>)}
                </tbody>
            </table>
        </div>}
    </div>;
}
