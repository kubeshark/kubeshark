import styles from "./EntrySections.module.sass";
import React, {useState} from "react";
import {SyntaxHighlighter} from "../UI/SyntaxHighlighter/index";
import CollapsibleContainer from "../UI/CollapsibleContainer";
import FancyTextDisplay from "../UI/FancyTextDisplay";
import Checkbox from "../UI/Checkbox";
import ProtobufDecoder from "protobuf-decoder";

interface EntryViewLineProps {
    label: string;
    value: number | string;
}

const EntryViewLine: React.FC<EntryViewLineProps> = ({label, value}) => {
    return (label && value && <tr className={styles.dataLine}>
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
            </tr>) || null;
}


interface EntrySectionCollapsibleTitleProps {
    title: string,
    color: string,
    isExpanded: boolean,
}

const EntrySectionCollapsibleTitle: React.FC<EntrySectionCollapsibleTitleProps> = ({title, color, isExpanded}) => {
    return <div className={styles.title}>
        <span className={`${styles.button} ${isExpanded ? styles.expanded : ''}`} style={{backgroundColor: color}}>
            {isExpanded ? '-' : '+'}
        </span>
        <span>{title}</span>
    </div>
}

interface EntrySectionContainerProps {
    title: string,
    color: string,
}

export const EntrySectionContainer: React.FC<EntrySectionContainerProps> = ({title, color, children}) => {
    const [expanded, setExpanded] = useState(true);
    return <CollapsibleContainer
        className={styles.collapsibleContainer}
        isExpanded={expanded}
        onClick={() => setExpanded(!expanded)}
        title={<EntrySectionCollapsibleTitle title={title} color={color} isExpanded={expanded}/>}
    >
        {children}
    </CollapsibleContainer>
}

interface EntryBodySectionProps {
    content: any,
    color: string,
    encoding?: string,
    contentType?: string,
}

export const EntryBodySection: React.FC<EntryBodySectionProps> = ({
    color,
    content,
    encoding,
    contentType,
}) => {
    const MAXIMUM_BYTES_TO_HIGHLIGHT = 10000; // The maximum of chars to highlight in body, in case the response can be megabytes
    const supportedLanguages = [['html', 'html'], ['json', 'json'], ['application/grpc', 'json']]; // [[indicator, languageToUse],...]
    const jsonLikeFormats = ['json'];
    const protobufFormats = ['application/grpc'];
    const [isWrapped, setIsWrapped] = useState(false);

    const formatTextBody = (body): string => {
        const chunk = body.slice(0, MAXIMUM_BYTES_TO_HIGHLIGHT);
        const bodyBuf = encoding === 'base64' ? atob(chunk) : chunk;

        try {
            if (jsonLikeFormats.some(format => contentType?.indexOf(format) > -1)) {
                return JSON.stringify(JSON.parse(bodyBuf), null, 2);
            } else if (protobufFormats.some(format => contentType?.indexOf(format) > -1)) {
                // Replace all non printable characters (ASCII)
                const protobufDecoder = new ProtobufDecoder(bodyBuf, true);
                return JSON.stringify(protobufDecoder.decode().toSimple(), null, 2);
            }
        } catch (error) {
            console.error(error);
        }
        return bodyBuf;
    }

    const getLanguage = (mimetype) => {
        const chunk = content?.slice(0, 100);
        if (chunk.indexOf('html') > 0 || chunk.indexOf('HTML') > 0) return supportedLanguages[0][1];
        const language = supportedLanguages.find(el => (mimetype + contentType).indexOf(el[0]) > -1);
        return language ? language[1] : 'default';
    }

    return <React.Fragment>
        {content && content?.length > 0 && <EntrySectionContainer title='Body' color={color}>
            <table>
                <tbody>
                    <EntryViewLine label={'Mime type'} value={contentType}/>
                    <EntryViewLine label={'Encoding'} value={encoding}/>
                </tbody>
            </table>

            <div style={{display: 'flex', alignItems: 'center', alignContent: 'center', margin: "5px 0"}} onClick={() => setIsWrapped(!isWrapped)}>
                <div style={{paddingTop: 3}}>
                    <Checkbox checked={isWrapped} onToggle={() => {}}/>
                </div>
                <span style={{marginLeft: '.5rem'}}>Wrap text</span>
            </div>

            <SyntaxHighlighter
                isWrapped={isWrapped}
                code={formatTextBody(content)}
                language={content?.mimeType ? getLanguage(content.mimeType) : 'default'}
            />
        </EntrySectionContainer>}
    </React.Fragment>
}

interface EntrySectionProps {
    title: string,
    color: string,
    arrayToIterate: any[],
}

export const EntryTableSection: React.FC<EntrySectionProps> = ({title, color, arrayToIterate}) => {
    return <React.Fragment>
        {
            arrayToIterate && arrayToIterate.length > 0 ?
                <EntrySectionContainer title={title} color={color}>
                    <table>
                        <tbody>
                            {arrayToIterate.map(({name, value}, index) => <EntryViewLine key={index} label={name}
                                                                                            value={value}/>)}
                        </tbody>
                    </table>
                </EntrySectionContainer> : <span/>
        }
    </React.Fragment>
}

interface EntryPolicySectionProps {
    title: string,
    color: string,
    latency?: number,
    arrayToIterate: any[],
}

interface EntryPolicySectionCollapsibleTitleProps {
    label: string;
    matched: string;
    isExpanded: boolean;
}

const EntryPolicySectionCollapsibleTitle: React.FC<EntryPolicySectionCollapsibleTitleProps> = ({label, matched, isExpanded}) => {
    return <div className={styles.title}>
        <span className={`${styles.button} ${isExpanded ? styles.expanded : ''}`}>
            {isExpanded ? '-' : '+'}
        </span>
        <span>
            <tr className={styles.dataLine}>
            <td className={`${styles.dataKey} ${styles.rulesTitleSuccess}`}>{label}</td>
            <td className={`${styles.dataKey} ${matched === 'Success' ? styles.rulesMatchedSuccess : styles.rulesMatchedFailure}`}>{matched}</td>
            </tr>
        </span>
    </div>
}

interface EntryPolicySectionContainerProps {
    label: string;
    matched: string;
    children?: any;
}

export const EntryPolicySectionContainer: React.FC<EntryPolicySectionContainerProps> = ({label, matched, children}) => {
    const [expanded, setExpanded] = useState(false);
    return <CollapsibleContainer
        className={styles.collapsibleContainer}
        isExpanded={expanded}
        onClick={() => setExpanded(!expanded)}
        title={<EntryPolicySectionCollapsibleTitle label={label} matched={matched} isExpanded={expanded}/>}
    >
        {children}
    </CollapsibleContainer>
}

export const EntryTablePolicySection: React.FC<EntryPolicySectionProps> = ({title, color, latency, arrayToIterate}) => {
    return <React.Fragment>
        {
            arrayToIterate && arrayToIterate.length > 0 ?
                <>
                <EntrySectionContainer title={title} color={color}>
                    <table>
                        <tbody>
                            {arrayToIterate.map(({rule, matched}, index) => {
                                    return (
                                        <EntryPolicySectionContainer key={index} label={rule.Name} matched={matched && (rule.Type === 'latency' ? rule.Latency >= latency : true)? "Success" : "Failure"}>
                                            {
                                                <>
                                                    {
                                                        rule.Key &&
                                                        <tr className={styles.dataValue}><td><b>Key:</b></td> <td>{rule.Key}</td></tr>
                                                    }
                                                    {
                                                        rule.Latency !== 0 &&
                                                        <tr className={styles.dataValue}><td><b>Latency:</b></td> <td>{rule.Latency}</td></tr>
                                                    }
                                                    {
                                                        rule.Method &&
                                                        <tr className={styles.dataValue}><td><b>Method:</b></td> <td>{rule.Method}</td></tr>
                                                    }
                                                    {
                                                        rule.Path &&
                                                        <tr className={styles.dataValue}><td><b>Path:</b></td> <td>{rule.Path}</td></tr>
                                                    }
                                                    {
                                                        rule.Service &&
                                                        <tr className={styles.dataValue}><td><b>Service:</b></td> <td>{rule.Service}</td></tr>
                                                    }
                                                    {
                                                        rule.Type &&
                                                        <tr className={styles.dataValue}><td><b>Type:</b></td> <td>{rule.Type}</td></tr>
                                                    }
                                                    {
                                                        rule.Value &&
                                                        <tr className={styles.dataValue}><td><b>Value:</b></td> <td>{rule.Value}</td></tr>
                                                    }
                                                </>
                                            }
                                        </EntryPolicySectionContainer>
                                    )
                                }
                            )
                            }
                        </tbody>
                    </table>
                </EntrySectionContainer>
                </> : <span className={styles.noRules}>No rules could be applied to this request.</span>
        }
    </React.Fragment>
}

interface EntryContractSectionProps {
    title: string,
    color: string,
    contractReason: string,
}

export const EntryContractSection: React.FC<EntryContractSectionProps> = ({title, color, contractReason}) => {
    return <React.Fragment>
        <EntrySectionContainer title={title} color={color}>
            {contractReason}
        </EntrySectionContainer>
    </React.Fragment>
}
