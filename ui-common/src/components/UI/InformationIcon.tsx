import React from "react";
import styles from "./style/InformationIcon.module.sass"

const DEFUALT_LINK = "https://getmizu.io/docs"

interface LinkProps {
    link?: string,
    className?: string
    title?: string
}

export const Link: React.FC<LinkProps> = ({ link, className, title, children }) => {
    return <a href={link} className={className} title={title} target="_blank">
        {children}
    </a>
}

export const InformationIcon: React.FC<LinkProps> = ({ className }) => {
    return <Link title="documentation" className={`${styles.linkStyle} ${className}`} link={DEFUALT_LINK}>
        <span>Docs</span>
    </Link>
}