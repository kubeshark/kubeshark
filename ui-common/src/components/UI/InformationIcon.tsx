import React from "react";
import infoImg from 'assets/info.svg';

const DEFUALT_LINK = "https://getmizu.io/docs"

export interface InformationIconProps{
    link?: string
}

export const InformationIcon: React.FC<InformationIconProps> = ({link}) => {

    return <React.Fragment>
        <a href={DEFUALT_LINK ? DEFUALT_LINK : link}>
            <img className="headerIcon"  src={infoImg} alt="Info icon"/>
        </a>
    </React.Fragment>
}


