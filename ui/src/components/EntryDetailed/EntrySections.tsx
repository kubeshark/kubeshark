import styles from "./EntrySections.module.sass";
import React, {useState} from "react";
import {SyntaxHighlighter} from "../UI/SyntaxHighlighter";
import CollapsibleContainer from "../UI/CollapsibleContainer";
import FancyTextDisplay from "../UI/FancyTextDisplay";
import Checkbox from "../UI/Checkbox";
import ProtobufDecoder from "protobuf-decoder";

interface ViewLineProps {
    label: string;
    value: number | string;
}

const ViewLine: React.FC<ViewLineProps> = ({label, value}) => {
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

interface SectionCollapsibleTitleProps {
    title: string;
    isExpanded: boolean;
}

const SectionCollapsibleTitle: React.FC<SectionCollapsibleTitleProps> = ({title, isExpanded}) => {
    return <div className={styles.title}>
        <span className={`${styles.button} ${isExpanded ? styles.expanded : ''}`}>
            {isExpanded ? '-' : '+'}
        </span>
        <span>{title}</span>
    </div>
}

interface SectionContainerProps {
    title: string;
}

export const SectionContainer: React.FC<SectionContainerProps> = ({title, children}) => {
    const [expanded, setExpanded] = useState(true);
    return <CollapsibleContainer
        className={styles.collapsibleContainer}
        isExpanded={expanded}
        onClick={() => setExpanded(!expanded)}
        title={<SectionCollapsibleTitle title={title} isExpanded={expanded}/>}
    >
        {children}
    </CollapsibleContainer>
}

interface BodySectionProps {
    content: any;
    encoding?: string;
    contentType?: string;
}

export const BodySection: React.FC<BodySectionProps> = ({content, encoding, contentType}) => {
    const MAXIMUM_BYTES_TO_HIGHLIGHT = 10000; // The maximum of chars to highlight in body, in case the response can be megabytes
    const supportedLanguages = [['html', 'html'], ['json', 'json'], ['application/grpc', 'json']]; // [[indicator, languageToUse],...]
    const jsonLikeFormats = ['json'];
    const protobufFormats = ['application/grpc'];
    const [isWrapped, setIsWrapped] = useState(false);

    const formatTextBody = (body): string => {
        const chunk = body.slice(0, MAXIMUM_BYTES_TO_HIGHLIGHT);
        const bodyBuf = encoding === 'base64' ? atob(chunk) : chunk;

        try {
            if (jsonLikeFormats.some(format => content?.mimeType?.indexOf(format) > -1)) {
                return JSON.stringify(JSON.parse(bodyBuf), null, 2);
            } else if (protobufFormats.some(format => content?.mimeType?.indexOf(format) > -1)) {
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
        const chunk = content.text?.slice(0, 100);
        if (chunk.indexOf('html') > 0 || chunk.indexOf('HTML') > 0) return supportedLanguages[0][1];
        const language = supportedLanguages.find(el => (mimetype + contentType).indexOf(el[0]) > -1);
        return language ? language[1] : 'default';
    }

    return <React.Fragment>
        {content && content.text?.length > 0 && <SectionContainer title='Body'>
            <table>
                <tbody>
                    <ViewLine label={'Mime type'} value={content?.mimeType}/>
                    <ViewLine label={'Encoding'} value={encoding}/>
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
                code={formatTextBody(content.text)}
                language={content?.mimeType ? getLanguage(content.mimeType) : 'default'}
            />
        </SectionContainer>}
    </React.Fragment>
}

interface TableSectionProps {
    title: string,
    arrayToIterate: any[],
}

export const TableSection: React.FC<TableSectionProps> = ({title, arrayToIterate}) => {
    return <React.Fragment>
        {
            arrayToIterate && arrayToIterate.length > 0 ?
                <SectionContainer title={title}>
                    <table>
                        <tbody>
                            {arrayToIterate.map(({name, value}, index) => <ViewLine key={index} label={name}
                                                                                            value={value}/>)}
                        </tbody>
                    </table>
                </SectionContainer> : <span/>
        }
    </React.Fragment>
}

interface HAREntryPolicySectionProps {
    service: string,
    title: string,
    response: any,
    latency?: number,
    arrayToIterate: any[],
}


interface HAREntryPolicySectionCollapsibleTitleProps {
    label: string;
    matched: string;
    isExpanded: boolean;
}

const HAREntryPolicySectionCollapsibleTitle: React.FC<HAREntryPolicySectionCollapsibleTitleProps> = ({label, matched, isExpanded}) => {
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

interface HAREntryPolicySectionContainerProps {
    label: string;
    matched: string;
    children?: any;
}

export const HAREntryPolicySectionContainer: React.FC<HAREntryPolicySectionContainerProps> = ({label, matched, children}) => {
    const [expanded, setExpanded] = useState(false);
    return <CollapsibleContainer
        className={styles.collapsibleContainer}
        isExpanded={expanded}
        onClick={() => setExpanded(!expanded)}
        title={<HAREntryPolicySectionCollapsibleTitle label={label} matched={matched} isExpanded={expanded}/>}
    >
        {children}
    </CollapsibleContainer>
}

export const HAREntryTablePolicySection: React.FC<HAREntryPolicySectionProps> = ({service, title, response, latency, arrayToIterate}) => {
    return <React.Fragment>
        {arrayToIterate && arrayToIterate.length > 0 ? <>
            <SectionContainer title={title}>
                <table>
                    <tbody>
                        {arrayToIterate.map(({rule, matched}, index) => {
                            return (<HAREntryPolicySectionContainer key={index} label={rule.Name} matched={matched && (rule.Type === 'latency' ? rule.Latency >= latency : true)? "Success" : "Failure"}>
                                    <>
                                        {rule.Key && <tr className={styles.dataValue}><td><b>Key</b>:</td><td>{rule.Key}</td></tr>}
                                        {rule.Latency && <tr className={styles.dataValue}><td><b>Latency:</b></td> <td>{rule.Latency}</td></tr>}
                                        {rule.Method && <tr className={styles.dataValue}><td><b>Method:</b></td> <td>{rule.Method}</td></tr>}
                                        {rule.Path && <tr className={styles.dataValue}><td><b>Path:</b></td> <td>{rule.Path}</td></tr>}
                                        {rule.Service && <tr className={styles.dataValue}><td><b>Service:</b></td> <td>{service}</td></tr>}
                                        {rule.Type && <tr className={styles.dataValue}><td><b>Type:</b></td> <td>{rule.Type}</td></tr>}
                                        {rule.Value && <tr className={styles.dataValue}><td><b>Value:</b></td> <td>{rule.Value}</td></tr>}
                                    </>
                                </HAREntryPolicySectionContainer>)})}
                    </tbody>
                </table>
            </SectionContainer>
        </> : <span className={styles.noRules}>No rules could be applied to this request.</span>}
    </React.Fragment>
}