import React, { CSSProperties } from "react";
import infoImg from 'assets/info.svg';
import styles from "./style/InformationIcon.module.sass"

const DEFUALT_LINK = "https://getmizu.io/docs"

export interface InformationIconProps{
    link?: string,
    style? : CSSProperties
}

export const InformationIcon: React.FC<InformationIconProps> = ({link,style}) => {
    return <React.Fragment>
        <a href={DEFUALT_LINK ? DEFUALT_LINK : link} style={style} className={styles.flex} title="documentation" target="_blank">
            <img className="headerIcon"  src={infoImg} alt="Info icon"/>
        </a>
    </React.Fragment>
}


