import styles from "./HAREntrySections.module.sass";
import React, {useState} from "react";
import {SyntaxHighlighter} from "../SyntaxHighlighter/index";
import CollapsibleContainer from "../CollapsibleContainer";
import FancyTextDisplay from "../FancyTextDisplay";
import Checkbox from "../Checkbox";

interface HAREntryViewLineProps {
    label: string;
    value: number | string;
}

const HAREntryViewLine: React.FC<HAREntryViewLineProps> = ({label, value}) => {
    return label && value && <tr className={styles.dataLine}>
                <td className={styles.dataKey}>{label}</td>
                <td>
                    <FancyTextDisplay
                        className={styles.dataValue}
                        text={value}
                        applyTextEllipsis={false}
                        flipped={true}
                        displayIconOnMouseOver={true}
                    />
                </td>
            </tr> || null;
}


interface HAREntrySectionCollapsibleTitleProps {
    title: string;
    isExpanded: boolean;
}

const HAREntrySectionCollapsibleTitle: React.FC<HAREntrySectionCollapsibleTitleProps> = ({title, isExpanded}) => {
    return <div className={styles.title}>
        <span className={`${styles.button} ${isExpanded ? styles.expanded : ''}`}>
            {isExpanded ? '-' : '+'}
        </span>
        <span>{title}</span>
    </div>
}

interface HAREntrySectionContainerProps {
    title: string;
}

export const HAREntrySectionContainer: React.FC<HAREntrySectionContainerProps> = ({title, children}) => {
    const [expanded, setExpanded] = useState(true);
    return <CollapsibleContainer
        className={styles.collapsibleContainer}
        isExpanded={expanded}
        onClick={() => setExpanded(!expanded)}
        title={<HAREntrySectionCollapsibleTitle title={title} isExpanded={expanded}/>}
    >
        {children}
    </CollapsibleContainer>
}

interface HAREntryBodySectionProps {
    content: any;
    encoding?: string;
    contentType?: string;
}

export const HAREntryBodySection: React.FC<HAREntryBodySectionProps> = ({
                                                                            content,
                                                                            encoding,
                                                                            contentType,
                                                                        }) => {
    const MAXIMUM_BYTES_TO_HIGHLIGHT = 10000; // The maximum of chars to highlight in body, in case the response can be megabytes
    const supportedLanguages = [['html', 'html'], ['json', 'json'], ['application/grpc', 'json']]; // [[indicator, languageToUse],...]
    const jsonLikeFormats = ['json', 'application/grpc'];
    const [isWrapped, setIsWrapped] = useState(false);

    const formatTextBody = (body): string => {
        const chunk = body.slice(0, MAXIMUM_BYTES_TO_HIGHLIGHT);
        const bodyBuf = encoding === 'base64' ? atob(chunk) : chunk;

        if (jsonLikeFormats.some(format => content?.mimeType?.indexOf(format) > -1)) {
            try {
                return JSON.stringify(JSON.parse(bodyBuf), null, 2);
            } catch (error) {
                console.error(error);
            }
        }
        return bodyBuf;
    }

    const getLanguage = (mimetype) => {
        const chunk = content.text?.slice(0, 100);
        if (chunk.indexOf('html') > 0 || chunk.indexOf('HTML') > 0) return supportedLanguages[0][1];
        const language = supportedLanguages.find(el => (mimetype + contentType).indexOf(el[0]) > -1);
        return language ? language[1] : 'default';
    }

    return <React.Fragment>
        {content && content.text?.length > 0 && <HAREntrySectionContainer title='Body'>
            <table>
                <tbody>
                    <HAREntryViewLine label={'Mime type'} value={content?.mimeType}/>
                    <HAREntryViewLine label={'Encoding'} value={encoding}/>
                </tbody>
            </table>

            <div style={{display: 'flex', alignItems: 'center', alignContent: 'center', margin: "5px 0"}} onClick={() => setIsWrapped(!isWrapped)}>
                <div style={{paddingTop: 3}}>
                    <Checkbox checked={isWrapped} onToggle={() => {}}/>
                </div>
                <span style={{marginLeft: '.5rem', color: "white"}}>Wrap text</span>
            </div>

            <SyntaxHighlighter
                isWrapped={isWrapped}
                code={formatTextBody(content.text)}
                language={content?.mimeType ? getLanguage(content.mimeType) : 'default'}
            />
        </HAREntrySectionContainer>}
    </React.Fragment>
}

interface HAREntrySectionProps {
    title: string,
    arrayToIterate: any[],
}

export const HAREntryTableSection: React.FC<HAREntrySectionProps> = ({title, arrayToIterate}) => {
    return <React.Fragment>
        {
            arrayToIterate && arrayToIterate.length > 0 ?
                <HAREntrySectionContainer title={title}>
                    <table>
                        <tbody>
                            {arrayToIterate.map(({name, value}, index) => <HAREntryViewLine key={index} label={name}
                                                                                            value={value}/>)}
                        </tbody>
                    </table>
                </HAREntrySectionContainer> : <span/>
        }
    </React.Fragment>
}
