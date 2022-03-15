import React from "react";
import collapsedImg from "assets/collapsed.svg";
import expandedImg from "assets/expanded.svg";
import styles from "./style/CollapsibleContainer.module.sass";

interface Props {
    title: string | React.ReactNode,
    expanded: boolean,
    titleClassName?: string,
    className?: string,
    stickyHeader?: boolean,
}

const CollapsibleContainer: React.FC<Props> = ({title, children, expanded, titleClassName, className, stickyHeader = false}) => {
    const classNames = `${styles.CollapsibleContainer} ${expanded ? `${styles.CollapsibleContainerExpanded}` : `${styles.CollapsibleContainerCollapsed}`} ${className ? className : ''}`;

    // This is needed to achieve the sticky header feature.
    // It is needed an un-contained component for the css to work properly.
    const content = <React.Fragment>
        <div
            className={`${styles.CollapsibleContainerHeader} ${stickyHeader ? `${styles.CollapsibleContainerHeaderSticky}` : ""}
                      ${expanded ? `${styles.CollapsibleContainerHeaderExpanded}` : ""}`}>
            {
                React.isValidElement(title)?
                    <React.Fragment>{title}</React.Fragment> :
                    <React.Fragment>
                        <div className={`${styles.CollapsibleContainerTitle} ${titleClassName ? titleClassName : ''}`}>{title}</div>
                        <img
                            className={styles.CollapsibleContainerExpandCollapseButton}
                            src={expanded ? expandedImg : collapsedImg}
                            alt="Expand/Collapse Button"
                        />
                    </React.Fragment>
            }
        </div>
        {expanded ? children : null}
    </React.Fragment>;

    return stickyHeader ? content : <div className={classNames}>{content}</div>;
};

export default CollapsibleContainer;
