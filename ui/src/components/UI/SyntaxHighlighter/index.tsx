import React, { useEffect, useState } from 'react';
import Lowlight from 'react-lowlight'
import 'highlight.js/styles/atom-one-light.css'
import styles from './index.module.sass';

import xml from 'highlight.js/lib/languages/xml'
import json from 'highlight.js/lib/languages/json'
import protobuf from 'highlight.js/lib/languages/protobuf'
import javascript from 'highlight.js/lib/languages/javascript'
import actionscript from 'highlight.js/lib/languages/actionscript'
import wasm from 'highlight.js/lib/languages/wasm'
import handlebars from 'highlight.js/lib/languages/handlebars'
import yaml from 'highlight.js/lib/languages/yaml'
import python from 'highlight.js/lib/languages/python'

Lowlight.registerLanguage('python', python);
Lowlight.registerLanguage('xml', xml);
Lowlight.registerLanguage('json', json);
Lowlight.registerLanguage('yaml', yaml);
Lowlight.registerLanguage('protobuf', protobuf);
Lowlight.registerLanguage('javascript', javascript);
Lowlight.registerLanguage('actionscript', actionscript);
Lowlight.registerLanguage('wasm', wasm);
Lowlight.registerLanguage('handlebars', handlebars);

interface Props {
    code: string;
    showLineNumbers?: boolean;
    language?: string;
}

export const SyntaxHighlighter: React.FC<Props> = ({
    code,
    showLineNumbers = false,
    language = null,
}) => {
    const [markers, setMarkers] = useState([])

    useEffect(() => {
        const newMarkers = code.split("\n").map((item, i) => {
            return {
                line: i + 1,
                className: styles.hljsMarkerLine
            }
        });
        setMarkers(showLineNumbers ? newMarkers : []);
    }, [showLineNumbers, code])

    return <div style={{ fontSize: ".75rem" }} className={styles.highlighterContainer}><Lowlight language={language ? language : ""} value={code} markers={markers} /></div>;
};

export default SyntaxHighlighter;
