import React, { useEffect, useState } from 'react';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import duplicateImg from "assets/duplicate.svg";
import styles from './style/FancyTextDisplay.module.sass';

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
    text = String(text);

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

    const textElement = <span className={styles.FancyTextDisplayText}>{text}</span>;

    const copyButton = isPossibleToCopy && text ? <CopyToClipboard text={text} onCopy={onCopy}>
                    <span
                        className={styles.FancyTextDisplayIcon}
                        title={`Copy "${text}" value to clipboard`}
                    >
                        <img src={duplicateImg} alt="Duplicate full value"/>
                        {showCopiedNotification && <span className={styles.FancyTextDisplayCopyNotifier}>Copied</span>}
                    </span>
				</CopyToClipboard> : null;

    return (
        <p
            className={`${styles.FancyTextDisplayContainer} ${className ? className : ''} ${displayIconOnMouseOver ? ` ${styles.displayIconOnMouseOver} ` : ''} ${applyTextEllipsis ? ` ${styles.FancyTextDisplayContainerEllipsis} `: ''}`}
            title={text}
            onMouseOver={ e => setShowTooltip(true)}
            onMouseLeave={ e => setShowTooltip(false)}
        >
            {!buttonOnly && flipped && textElement}
            {copyButton}
            {!buttonOnly && !flipped && textElement}
            {useTooltip && showTooltip && <span className={styles.FancyTextDisplayCopyNotifier}>{text}</span>}
        </p>
    );
};

export default FancyTextDisplay;
