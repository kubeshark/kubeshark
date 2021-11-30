import React, { useEffect, useState } from 'react';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import AddCircleIcon from '@material-ui/icons/AddCircle';
import './style/Queryable.sass';

interface Props {
    query: string,
    updateQuery: any,
    style?: object,
    iconStyle?: object,
    className?: string,
    useTooltip?: boolean,
    displayIconOnMouseOver?: boolean,
    flipped?: boolean,
}

const Queryable: React.FC<Props> = ({query, updateQuery, style, iconStyle, className, useTooltip= true, displayIconOnMouseOver = false, flipped = false, children}) => {
    const [showAddedNotification, setAdded] = useState(false);
    const [showTooltip, setShowTooltip] = useState(false);

    const onCopy = () => {
        setAdded(true)
    };

    useEffect(() => {
        let timer;
        if (showAddedNotification) {
            updateQuery(query);
            timer = setTimeout(() => {
                setAdded(false);
            }, 1000);
        }
        return () => clearTimeout(timer);
    }, [showAddedNotification, query, updateQuery]);

    const addButton = query ? <CopyToClipboard text={query} onCopy={onCopy}>
                    <span
                        className={`Queryable-Icon`}
                        title={`Add "${query}" to the filter`}
                        style={iconStyle}
                    >
                        <AddCircleIcon fontSize="small" color="inherit"></AddCircleIcon>
                        {showAddedNotification && <span className={'Queryable-AddNotifier'}>Added</span>}
                    </span>
				</CopyToClipboard> : null;

    return (
        <div
            className={`Queryable-Container displayIconOnMouseOver ${className ? className : ''} ${displayIconOnMouseOver ? 'displayIconOnMouseOver ' : ''}`}
            style={style}
            onMouseOver={ e => setShowTooltip(true)}
            onMouseLeave={ e => setShowTooltip(false)}
        >
            {flipped && addButton}
            {children}
            {!flipped && addButton}
            {useTooltip && showTooltip && <span className={'Queryable-Tooltip'}>{query}</span>}
        </div>
    );
};

export default Queryable;
