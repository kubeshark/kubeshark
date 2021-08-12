import React, {useState} from "react";
import collapsedImg from "../assets/collapsed.svg";
import expandedImg from "../assets/expanded.svg";
import "./style/CollapsibleContainer.sass";

interface Props {
    title: string | React.ReactNode,
    onClick?: (e: React.SyntheticEvent) => void,
    isExpanded?: boolean,
    titleClassName?: string,
    stickyHeader?: boolean,
    className?: string,
    initialExpanded?: boolean;
    passiveOnClick?: boolean; //whether specifying onClick overrides internal _isExpanded state handling
}

const CollapsibleContainer: React.FC<Props> = ({title, children, isExpanded, onClick, titleClassName, stickyHeader = false, className, initialExpanded = true, passiveOnClick}) => {
    const [_isExpanded, _setExpanded] = useState(initialExpanded);
    let expanded = isExpanded !== undefined ? isExpanded : _isExpanded;
    const classNames = `CollapsibleContainer ${expanded ? "CollapsibleContainer-Expanded" : "CollapsibleContainer-Collapsed"} ${className ? className : ''}`;

    // This is needed to achieve the sticky header feature. 
    // It is needed an un-contained component for the css to work properly.
    const content = <React.Fragment>
        <div
            className={`CollapsibleContainer-Header ${stickyHeader ? "CollapsibleContainer-Header-Sticky" : ""} 
            ${expanded ? "CollapsibleContainer-Header-Expanded" : ""}`}
            onClick={(e) => {
                if (onClick) {
                    onClick(e)
                    if (passiveOnClick !== true)
                        return;
                }
                _setExpanded(!_isExpanded)
            }}
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
