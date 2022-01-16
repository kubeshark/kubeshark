import React from "react";
import collapsedImg from "../assets/collapsed.svg";
import expandedImg from "../assets/expanded.svg";
import "./style/CollapsibleContainer.sass";

interface Props {
    title: string | React.ReactNode,
    expanded: boolean,
    titleClassName?: string,
    className?: string,
    stickyHeader?: boolean,
}

const CollapsibleContainer: React.FC<Props> = ({title, children, expanded, titleClassName, className, stickyHeader = false}) => {
    const classNames = `CollapsibleContainer ${expanded ? "CollapsibleContainer-Expanded" : "CollapsibleContainer-Collapsed"} ${className ? className : ''}`;

    // This is needed to achieve the sticky header feature.
    // It is needed an un-contained component for the css to work properly.
    const content = <React.Fragment>
        <div
            className={`CollapsibleContainer-Header ${stickyHeader ? "CollapsibleContainer-Header-Sticky" : ""}
            ${expanded ? "CollapsibleContainer-Header-Expanded" : ""}`}
        >
            {
                React.isValidElement(title)?
                    <React.Fragment>{title}</React.Fragment> :
                    <React.Fragment>
                        <div className={`CollapsibleContainer-Title ${titleClassName ? titleClassName : ''}`}>{title}</div>
                        <img
                            className="CollapsibleContainer-ExpandCollapseButton"
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
