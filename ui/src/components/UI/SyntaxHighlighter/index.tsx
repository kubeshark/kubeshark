import React from 'react';
import './index.scss';
import Lowlight from 'react-lowlight'
import 'highlight.js/styles/atom-one-light.css'

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
                                                showLineNumbers = true,
                                                language = null
                                            }) => {
    return <Lowlight language={language ? language : ""} value={code} />;
};

export default SyntaxHighlighter;
