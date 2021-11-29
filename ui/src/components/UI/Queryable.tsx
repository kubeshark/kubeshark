import React, { useEffect, useState } from 'react';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import AddCircleIcon from '@material-ui/icons/AddCircle';
import './style/Queryable.sass';

interface Props {
    text: string | number,
    query: string,
    updateQuery: any,
    title?: string,
    textStyle?: object,
    wrapperStyle?: object,
    className?: string,
    isPossibleToCopy?: boolean,
    applyTextEllipsis?: boolean,
    useTooltip?: boolean,
    displayIconOnMouseOver?: boolean,
    onClick?: React.EventHandler<React.MouseEvent<HTMLElement>>;
}

const Queryable: React.FC<Props> = ({text, query, updateQuery, title, textStyle, wrapperStyle, className, isPossibleToCopy = true, applyTextEllipsis = true, useTooltip= false, displayIconOnMouseOver = false}) => {
    const [showAddedNotification, setAdded] = useState(false);
    const [showTooltip, setShowTooltip] = useState(false);
    text = String(text);

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
        // eslint-disable-next-line
    }, [showAddedNotification]);

    const textElement = <span title={title} className={'Queryable-Text'} style={textStyle}>{text}</span>;

    const copyButton = text ? <CopyToClipboard text={text} onCopy={onCopy}>
                    <span
                        className={`Queryable-Icon`}
                        title={`Add "${query}" to the filter`}
                    >
                        <AddCircleIcon fontSize="small" color="inherit"></AddCircleIcon>
                        {showAddedNotification && <span className={'Queryable-CopyNotifier'}>Added</span>}
                    </span>
				</CopyToClipboard> : null;

    return (
        <div
            className={`Queryable-Container displayIconOnMouseOver ${className ? className : ''} ${displayIconOnMouseOver ? 'displayIconOnMouseOver ' : ''} ${applyTextEllipsis ? ' Queryable-ContainerEllipsis' : ''}`}
            style={wrapperStyle}
            onMouseOver={ e => setShowTooltip(true)}
            onMouseLeave={ e => setShowTooltip(false)}
        >
            {textElement}
            {copyButton}
            {useTooltip && showTooltip && <span className={'Queryable-CopyNotifier'}>{text}</span>}
        </div>
    );
};

export default Queryable;
