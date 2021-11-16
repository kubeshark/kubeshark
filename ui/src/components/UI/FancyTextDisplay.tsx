import React, { useEffect, useState } from 'react';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import duplicateImg from "../assets/duplicate.svg";
import './style/FancyTextDisplay.sass';

interface Props {
    text: string | number,
    className?: string,
    isPossibleToCopy?: boolean,
    applyTextEllipsis?: boolean,
    flipped?: boolean,
    useTooltip?: boolean,
    displayIconOnMouseOver?: boolean,
    buttonOnly?: boolean,
}

const FancyTextDisplay: React.FC<Props> = ({text, className, isPossibleToCopy = true, applyTextEllipsis = true, flipped = false, useTooltip= false, displayIconOnMouseOver = false, buttonOnly = false}) => {
    const [showCopiedNotification, setCopied] = useState(false);
    const [showTooltip, setShowTooltip] = useState(false);

    const onCopy = () => {
        setCopied(true)
    };

    useEffect(() => {
        let timer;
        if (showCopiedNotification) {
            timer = setTimeout(() => {
                setCopied(false);
            }, 1000);
        }
        return () => clearTimeout(timer);
    }, [showCopiedNotification]);

    const textElement = <span className={'FancyTextDisplay-Text'}>{text}</span>;

    const copyButton = isPossibleToCopy && text ? <CopyToClipboard text={text} onCopy={onCopy}>
                    <span
                        className={`FancyTextDisplay-Icon`}
                        title={`Copy "${text}" value to clipboard`}
                    >
                        <img src={duplicateImg} alt="Duplicate full value"/>
                        {showCopiedNotification && <span className={'FancyTextDisplay-CopyNotifier'}>Copied</span>}
                    </span>
				</CopyToClipboard> : null;

    return (
        <p
            className={`FancyTextDisplay-Container ${className ? className : ''} ${displayIconOnMouseOver ? 'displayIconOnMouseOver ' : ''} ${applyTextEllipsis ? ' FancyTextDisplay-ContainerEllipsis' : ''}`}
            title={text.toString()}
            onMouseOver={ e => setShowTooltip(true)}
            onMouseLeave={ e => setShowTooltip(false)}
        >
            {!buttonOnly && flipped && textElement}
            {copyButton}
            {!buttonOnly && !flipped && textElement}
            {useTooltip && showTooltip && <span className={'FancyTextDisplay-CopyNotifier'}>{text}</span>}
        </p>
    );
};

export default FancyTextDisplay;
