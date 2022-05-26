import React, { useEffect, useState } from 'react';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import AddCircleIcon from '@material-ui/icons/AddCircle';
import QueryableStyle from './style/Queryable.module.sass';
import {useRecoilState} from "recoil";
import queryAtom from "../../recoil/query";

interface Props {
    query: string,
    style?: object,
    iconStyle?: object,
    className?: string,
    useTooltip?: boolean,
    tooltipStyle?: object,
    displayIconOnMouseOver?: boolean,
    flipped?: boolean,
}

const Queryable: React.FC<Props> = ({query, style, iconStyle, className, useTooltip = true, tooltipStyle = null, displayIconOnMouseOver = false, flipped = false, children}) => {
    const [showAddedNotification, setAdded] = useState(false);
    const [showTooltip, setShowTooltip] = useState(false);
    const [queryState, setQuery] = useRecoilState(queryAtom);

    const onCopy = () => {
        setAdded(true)
    };

    useEffect(() => {
        let timer;
        if (showAddedNotification) {
            setQuery(queryState ? `${queryState} and ${query}` : query);
            timer = setTimeout(() => {
                setAdded(false);
            }, 1000);
        }
        return () => clearTimeout(timer);

    // eslint-disable-next-line
    }, [showAddedNotification, query, setQuery]);

    const addButton = query ? <CopyToClipboard text={query} onCopy={onCopy}>
                    <span
                        className={QueryableStyle.QueryableIcon}
                        title={`Add "${query}" to the filter`}
                        style={iconStyle}>
                        <AddCircleIcon fontSize="small" color="inherit"/>
                        {showAddedNotification && <span className={QueryableStyle.QueryableAddNotifier}>Added</span>}
                    </span>
				</CopyToClipboard> : null;

    return (
        <div className={`${QueryableStyle.QueryableContainer} ${QueryableStyle.displayIconOnMouseOver} ${className ? className : ''} ${displayIconOnMouseOver ? QueryableStyle.displayIconOnMouseOver : ''}`}
            style={style} onMouseOver={ e => setShowTooltip(true)} onMouseLeave={ e => setShowTooltip(false)}>
                {flipped && addButton}
                {children}
                {!flipped && addButton}
                {useTooltip && showTooltip && (query !== "") && <span data-cy={"QueryableTooltip"} className={QueryableStyle.QueryableTooltip} style={tooltipStyle}>{query}</span>}
        </div>
    );
};

export default Queryable;
