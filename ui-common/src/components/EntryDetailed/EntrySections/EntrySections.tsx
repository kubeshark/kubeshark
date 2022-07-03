import styles from "./EntrySections.module.sass";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import { SyntaxHighlighter } from "../../UI/SyntaxHighlighter";
import CollapsibleContainer from "../../UI/CollapsibleContainer/CollapsibleContainer";
import FancyTextDisplay from "../../UI/FancyTextDisplay/FancyTextDisplay";
import Queryable from "../../UI/Queryable/Queryable";
import Checkbox from "../../UI/Checkbox/Checkbox";
import ProtobufDecoder from "protobuf-decoder";
import { default as jsonBeautify } from "json-beautify";
import { default as xmlBeautify } from "xml-formatter";
import { Utils } from "../../../helpers/Utils"

interface EntryViewLineProps {
    label: string;
    value: number | string;
    selector?: string;
    overrideQueryValue?: string;
    displayIconOnMouseOver?: boolean;
    useTooltip?: boolean;
}

const EntryViewLine: React.FC<EntryViewLineProps> = ({ label, value, selector = "", overrideQueryValue = "", displayIconOnMouseOver = true, useTooltip = true }) => {
    let query: string;
    if (!selector) {
        query = "";
    } else if (overrideQueryValue) {
        query = `${selector} == ${overrideQueryValue}`;
    } else if (typeof (value) == "string") {
        query = `${selector} == "${JSON.stringify(value).slice(1, -1)}"`;
    } else {
        query = `${selector} == ${value}`;
    }
    return (label && <tr className={styles.dataLine}>
        <td className={`${styles.dataKey}`}>
            <Queryable
                query={query}
                style={{ float: "right", height: "18px" }}
                iconStyle={{ marginRight: "20px" }}
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
}

const EntrySectionCollapsibleTitle: React.FC<EntrySectionCollapsibleTitleProps> = ({ title, color, expanded, setExpanded, query = "" }) => {
    return <div className={styles.title}>
        <div
            className={`${styles.button} ${expanded ? styles.expanded : ''}`}
            style={{ backgroundColor: color }}
            onClick={() => {
                setExpanded(!expanded)
            }}
        >
            {expanded ? '-' : '+'}
        </div>
        <Queryable
            query={query}
            useTooltip={!!query}
            displayIconOnMouseOver={!!query}
        >
            <span>{title}</span>
        </Queryable>
    </div>
}

interface EntrySectionContainerProps {
    title: string,
    color: string,
    query?: string,
}

export const EntrySectionContainer: React.FC<EntrySectionContainerProps> = ({ title, color, children, query = "" }) => {
    const [expanded, setExpanded] = useState(true);
    return <CollapsibleContainer
        className={styles.collapsibleContainer}
        expanded={expanded}
        title={<EntrySectionCollapsibleTitle title={title} color={color} expanded={expanded} setExpanded={setExpanded} query={query} />}
    >
        {children}
    </CollapsibleContainer>
}

const MAXIMUM_BYTES_TO_FORMAT = 1000000; // The maximum of chars to highlight in body, in case the response can be megabytes
const jsonLikeFormats = ['json', 'yaml', 'yml'];
const xmlLikeFormats = ['xml', 'html'];
const protobufFormats = ['application/grpc'];
const supportedFormats = jsonLikeFormats.concat(xmlLikeFormats, protobufFormats);

interface EntryBodySectionProps {
    title: string,
    content: any,
    color: string,
    encoding?: string,
    contentType?: string,
    selector?: string,
}

export const formatRequest = (body: any, contentType: string, decodeBase64: boolean = true, isBase64Encoding: boolean = false, isPretty: boolean = true): string => {
    if (!decodeBase64 || !body) return body;

    const chunk = body.slice(0, MAXIMUM_BYTES_TO_FORMAT);
    const bodyBuf = isBase64Encoding ? atob(chunk) : chunk;

    try {
        if (jsonLikeFormats.some(format => contentType?.indexOf(format) > -1)) {
            if (!isPretty) return bodyBuf;
            return Utils.isJson(bodyBuf) ? jsonBeautify(JSON.parse(bodyBuf), null, 2, 80) : bodyBuf
        } else if (xmlLikeFormats.some(format => contentType?.indexOf(format) > -1)) {
            if (!isPretty) return bodyBuf;
            return xmlBeautify(bodyBuf, {
                indentation: '  ',
                filter: (node) => node.type !== 'Comment',
                collapseContent: true,
                lineSeparator: '\n'
            });
        } else if (protobufFormats.some(format => contentType?.indexOf(format) > -1)) {
            // Replace all non printable characters (ASCII)
            const protobufDecoder = new ProtobufDecoder(bodyBuf, true);
            const protobufDecoded = protobufDecoder.decode().toSimple();
            if (!isPretty) return JSON.stringify(protobufDecoded);
            return jsonBeautify(protobufDecoded, null, 2, 80);
        }
    } catch (error) {
        console.error(error)
        throw error
    }

    return bodyBuf;
}

export const EntryBodySection: React.FC<EntryBodySectionProps> = ({
    title,
    color,
    content,
    encoding,
    contentType,
    selector,
}) => {
    const [isPretty, setIsPretty] = useState(true);
    const [showLineNumbers, setShowLineNumbers] = useState(false);
    const [decodeBase64, setDecodeBase64] = useState(true);

    const isBase64Encoding = encoding === 'base64';
    const supportsPrettying = supportedFormats.some(format => contentType?.indexOf(format) > -1);
    const [isDecodeGrpc, setIsDecodeGrpc] = useState(true);
    const [isLineNumbersGreaterThenOne, setIsLineNumbersGreaterThenOne] = useState(true);

    useEffect(() => {
        (isLineNumbersGreaterThenOne && isPretty) && setShowLineNumbers(true);
        !isLineNumbersGreaterThenOne && setShowLineNumbers(false);
    }, [isLineNumbersGreaterThenOne, isPretty])

    const formatTextBody = useCallback((body) => {
        try {
            return formatRequest(body, contentType, decodeBase64, isBase64Encoding, isPretty)
        } catch (error) {
            if (String(error).includes("More than one message in")) {
                if (isDecodeGrpc)
                    setIsDecodeGrpc(false);
            } else if (String(error).includes("Failed to parse")) {
                console.warn(error);
            }
        }
    }, [isPretty, contentType, isDecodeGrpc, decodeBase64, isBase64Encoding])

    const formattedText = useMemo(() => formatTextBody(content), [formatTextBody, content]);

    useEffect(() => {
        const lineNumbers = Utils.lineNumbersInString(formattedText);
        setIsLineNumbersGreaterThenOne(lineNumbers > 1);
    }, [isPretty, content, showLineNumbers, formattedText]);

    return <React.Fragment>
        {content && content?.length > 0 && <EntrySectionContainer
            title={title}
            color={color}
            query={`${selector} == r".*"`}
        >
            <div style={{ display: 'flex', alignItems: 'center', alignContent: 'center', margin: "5px 0" }}>
                {supportsPrettying && <div style={{ paddingTop: 3 }}>
                    <Checkbox checked={isPretty} onToggle={() => { setIsPretty(!isPretty) }} data-cy="prettyCheckBoxInput"/>
                </div>}
                {supportsPrettying && <span style={{ marginLeft: '.2rem' }}>Pretty</span>}

                <div style={{ paddingTop: 3, paddingLeft: supportsPrettying ? 20 : 0 }}>
                    <Checkbox checked={showLineNumbers} onToggle={() => { setShowLineNumbers(!showLineNumbers) }} disabled={!isLineNumbersGreaterThenOne || !decodeBase64} data-cy="lineNumbersCheckBoxInput"/>
                </div>
                <span style={{ marginLeft: '.2rem' }}>Line numbers</span>

                {isBase64Encoding && <div style={{ paddingTop: 3, paddingLeft: (isLineNumbersGreaterThenOne || supportsPrettying) ? 20 : 0 }}>
                    <Checkbox checked={decodeBase64} onToggle={() => { setDecodeBase64(!decodeBase64) }}  data-cy="decodeBase64CheckboxInput"/>
                </div>}
                {isBase64Encoding && <span style={{ marginLeft: '.2rem' }}>Decode Base64</span>}
                {!isDecodeGrpc && <span style={{ fontSize: '12px', color: '#DB2156', marginLeft: '.8rem' }}>More than one message in protobuf payload is not supported</span>}
            </div>

            <SyntaxHighlighter
                code={formattedText}
                showLineNumbers={showLineNumbers}
            />

        </EntrySectionContainer>}
    </React.Fragment>
}


interface EntrySectionProps {
    title: string,
    color: string,
    arrayToIterate: any[],
}

export const EntryTableSection: React.FC<EntrySectionProps> = ({ title, color, arrayToIterate }) => {
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
                            {arrayToIterateSorted.map(({ name, value, selector }, index) => <EntryViewLine
                                key={index}
                                label={name}
                                value={value}
                                selector={selector}
                            />)}
                        </tbody>
                    </table>
                </EntrySectionContainer> : <span />
        }
    </React.Fragment>
}
