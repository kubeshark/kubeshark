import React, { CSSProperties } from "react";
import styles from "./style/InformationIcon.module.sass"

const DEFUALT_LINK = "https://getmizu.io/docs"

export interface InformationIconProps {
    link?: string,
    style?: CSSProperties
}

export const InformationIcon: React.FC<InformationIconProps> = ({ link, style }) => {
    return <React.Fragment>
        <a href={DEFUALT_LINK ? DEFUALT_LINK : link} style={style} className={styles.linkStyle} title="documentation" target="_blank">
            <span>Docs</span>
        </a>
    </React.Fragment>
}


