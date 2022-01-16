import styles from "./EntrySections.module.sass";
import React, {useState} from "react";
import {SyntaxHighlighter} from "../UI/SyntaxHighlighter/index";
import CollapsibleContainer from "../UI/CollapsibleContainer";
import FancyTextDisplay from "../UI/FancyTextDisplay";
import Queryable from "../UI/Queryable";
import Checkbox from "../UI/Checkbox";
import ProtobufDecoder from "protobuf-decoder";
import {default as jsonBeautify} from "json-beautify";
import {default as xmlBeautify} from "xml-formatter";

interface EntryViewLineProps {
    label: string;
    value: number | string;
    updateQuery?: any;
    selector?: string;
    overrideQueryValue?: string;
    displayIconOnMouseOver?: boolean;
    useTooltip?: boolean;
}

const EntryViewLine: React.FC<EntryViewLineProps> = ({label, value, updateQuery = null, selector = "", overrideQueryValue = "", displayIconOnMouseOver = true, useTooltip = true}) => {
    let query: string;
    if (!selector) {
        query = "";
    } else if (overrideQueryValue) {
        query = `${selector} == ${overrideQueryValue}`;
    } else if (typeof(value) == "string") {
        query = `${selector} == "${JSON.stringify(value).slice(1, -1)}"`;
    } else {
        query = `${selector} == ${value}`;
    }
    return (label && <tr className={styles.dataLine}>
                    <td className={`${styles.dataKey}`}>
                        <Queryable
                            query={query}
                            updateQuery={updateQuery}
                            style={{float: "right", height: "18px"}}
                            iconStyle={{marginRight: "20px"}}
                            flipped={true}
                            useTooltip={useTooltip}
                            displayIconOnMouseOver={displayIconOnMouseOver}
                        >
                            {label}
                        </Queryable>
                    </td>
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
    expanded: boolean,
    setExpanded: any,
    query?: string,
    updateQuery?: any,
}

const EntrySectionCollapsibleTitle: React.FC<EntrySectionCollapsibleTitleProps> = ({title, color, expanded, setExpanded, query = "", updateQuery = null}) => {
    return <div className={styles.title}>
        <div
            className={`${styles.button} ${expanded ? styles.expanded : ''}`}
            style={{backgroundColor: color}}
            onClick={() => {
                setExpanded(!expanded)
            }}
        >
            {expanded ? '-' : '+'}
        </div>
        <Queryable
            query={query}
            updateQuery={updateQuery}
            useTooltip={updateQuery ? true : false}
            displayIconOnMouseOver={updateQuery ? true : false}
        >
            <span>{title}</span>
        </Queryable>
    </div>
}

interface EntrySectionContainerProps {
    title: string,
    color: string,
    query?: string,
    updateQuery?: any,
}

export const EntrySectionContainer: React.FC<EntrySectionContainerProps> = ({title, color, children, query = "", updateQuery = null}) => {
    const [expanded, setExpanded] = useState(true);
    return <CollapsibleContainer
        className={styles.collapsibleContainer}
        expanded={expanded}
        title={<EntrySectionCollapsibleTitle title={title} color={color} expanded={expanded} setExpanded={setExpanded} query={query} updateQuery={updateQuery}/>}
    >
        {children}
    </CollapsibleContainer>
}

interface EntryBodySectionProps {
    title: string,
    content: any,
    color: string,
    updateQuery: any,
    encoding?: string,
    contentType?: string,
    selector?: string,
}

export const EntryBodySection: React.FC<EntryBodySectionProps> = ({
    title,
    color,
    updateQuery,
    content,
    encoding,
    contentType,
    selector,
}) => {
    const MAXIMUM_BYTES_TO_FORMAT = 1000000; // The maximum of chars to highlight in body, in case the response can be megabytes
    const jsonLikeFormats = ['json', 'yaml', 'yml'];
    const xmlLikeFormats = ['xml', 'html'];
    const protobufFormats = ['application/grpc'];
    const supportedFormats = jsonLikeFormats.concat(xmlLikeFormats, protobufFormats);

    const [isPretty, setIsPretty] = useState(true);
    const [showLineNumbers, setShowLineNumbers] = useState(true);
    const [decodeBase64, setDecodeBase64] = useState(true);

    const isBase64Encoding = encoding === 'base64';
    const supportsPrettying = supportedFormats.some(format => contentType?.indexOf(format) > -1);

    const formatTextBody = (body: any): string => {
        if (!decodeBase64) return body;

        const chunk = body.slice(0, MAXIMUM_BYTES_TO_FORMAT);
        const bodyBuf = isBase64Encoding ? atob(chunk) : chunk;

        if (!isPretty) return bodyBuf;

        try {
            if (jsonLikeFormats.some(format => contentType?.indexOf(format) > -1)) {
                return jsonBeautify(JSON.parse(bodyBuf), null, 2, 80);
            }  else if (xmlLikeFormats.some(format => contentType?.indexOf(format) > -1)) {
                return xmlBeautify(bodyBuf, {
                    indentation: '  ',
                    filter: (node) => node.type !== 'Comment',
                    collapseContent: true,
                    lineSeparator: '\n'
                });
            } else if (protobufFormats.some(format => contentType?.indexOf(format) > -1)) {
                // Replace all non printable characters (ASCII)
                const protobufDecoder = new ProtobufDecoder(bodyBuf, true);
                return jsonBeautify(protobufDecoder.decode().toSimple(), null, 2, 80);
            }
        } catch (error) {
            console.error(error);
        }
        return bodyBuf;
    }

    return <React.Fragment>
        {content && content?.length > 0 && <EntrySectionContainer
                                                title={title}
                                                color={color}
                                                query={`${selector} == r".*"`}
                                                updateQuery={updateQuery}
                                            >
            <div style={{display: 'flex', alignItems: 'center', alignContent: 'center', margin: "5px 0"}}>
                {supportsPrettying && <div style={{paddingTop: 3}}>
                    <Checkbox checked={isPretty} onToggle={() => {setIsPretty(!isPretty)}}/>
                </div>}
                {supportsPrettying && <span style={{marginLeft: '.2rem'}}>Pretty</span>}

                <div style={{paddingTop: 3, paddingLeft: supportsPrettying ? 20 : 0}}>
                    <Checkbox checked={showLineNumbers} onToggle={() => {setShowLineNumbers(!showLineNumbers)}}/>
                </div>
                <span style={{marginLeft: '.2rem'}}>Line numbers</span>

                {isBase64Encoding && <div style={{paddingTop: 3, paddingLeft: 20}}>
                    <Checkbox checked={decodeBase64} onToggle={() => {setDecodeBase64(!decodeBase64)}}/>
                </div>}
                {isBase64Encoding && <span style={{marginLeft: '.2rem'}}>Decode Base64</span>}
            </div>

            <SyntaxHighlighter
                code={formatTextBody(content)}
                showLineNumbers={showLineNumbers}
            />
        </EntrySectionContainer>}
    </React.Fragment>
}

interface EntrySectionProps {
    title: string,
    color: string,
    arrayToIterate: any[],
    updateQuery: any,
}

export const EntryTableSection: React.FC<EntrySectionProps> = ({title, color, arrayToIterate, updateQuery}) => {
    let arrayToIterateSorted: any[];
    if (arrayToIterate) {
        arrayToIterateSorted = arrayToIterate.sort((a, b) => {
            if (a.name > b.name) {
                return 1;
            }

            if (a.name < b.name) {
                return -1;
            }

            return 0;
        });
    }
    return <React.Fragment>
        {
            arrayToIterate && arrayToIterate.length > 0 ?
                <EntrySectionContainer title={title} color={color}>
                    <table>
                        <tbody id={`tbody-${title}`}>
                            {arrayToIterateSorted.map(({name, value, selector}, index) => <EntryViewLine
                                key={index}
                                label={name}
                                value={value}
                                updateQuery={updateQuery}
                                selector={selector}
                            />)}
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
    expanded: boolean;
    setExpanded: any;
}

const EntryPolicySectionCollapsibleTitle: React.FC<EntryPolicySectionCollapsibleTitleProps> = ({label, matched, expanded, setExpanded}) => {
    return <div className={styles.title}>
        <span
            className={`${styles.button}
            ${expanded ? styles.expanded : ''}`}
            onClick={() => {
                setExpanded(!expanded)
            }}
        >
            {expanded ? '-' : '+'}
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
        expanded={expanded}
        title={<EntryPolicySectionCollapsibleTitle label={label} matched={matched} expanded={expanded} setExpanded={setExpanded}/>}
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
                                        <EntryPolicySectionContainer key={index} label={rule.Name} matched={matched && (rule.Type === 'slo' ? rule.ResponseTime >= latency : true)? "Success" : "Failure"}>
                                            {
                                                <>
                                                    {
                                                        rule.Key &&
                                                        <tr className={styles.dataValue}><td><b>Key:</b></td> <td>{rule.Key}</td></tr>
                                                    }
                                                    {
                                                        rule.ResponseTime !== 0 &&
                                                        <tr className={styles.dataValue}><td><b>Response Time:</b></td> <td>{rule.ResponseTime}</td></tr>
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
    color: string,
    requestReason: string,
    responseReason: string,
    contractContent: string,
}

export const EntryContractSection: React.FC<EntryContractSectionProps> = ({color, requestReason, responseReason, contractContent}) => {
    return <React.Fragment>
        {requestReason && <EntrySectionContainer title="Request" color={color}>
            {requestReason}
        </EntrySectionContainer>}
        {responseReason && <EntrySectionContainer title="Response" color={color}>
            {responseReason}
        </EntrySectionContainer>}
        {contractContent && <EntrySectionContainer title="Contract" color={color}>
            <SyntaxHighlighter
                code={contractContent}
                language={"yaml"}
            />
        </EntrySectionContainer>}
    </React.Fragment>
}
