import React from "react";
import AceEditor from "react-ace";
import { config } from 'ace-builds';

import "ace-builds/src-noconflict/ext-searchbox";
import "ace-builds/src-noconflict/mode-python";
import "ace-builds/src-noconflict/mode-json";
import "ace-builds/src-noconflict/theme-github";
import "ace-builds/src-noconflict/mode-javascript";
import "ace-builds/src-noconflict/mode-xml";
import "ace-builds/src-noconflict/mode-html";



config.set(
    "basePath",
    "https://cdn.jsdelivr.net/npm/ace-builds@1.4.6/src-noconflict/"
);
config.setModuleUrl(
    "ace/mode/javascript_worker",
    "https://cdn.jsdelivr.net/npm/ace-builds@1.4.6/src-noconflict/worker-javascript.js"
);

export interface CodeEditorProps {
    code: string,
    onChange?: (code: string) => void,
    language?: string
}
const CodeEditor: React.FC<CodeEditorProps> = ({
    language,
    onChange,
    code
}) => {
    return (
        <AceEditor
            mode={language}
            theme="github"
            onChange={onChange}
            editorProps={{ $blockScrolling: true }}
            setOptions={{
                enableBasicAutocompletion: true,
                enableLiveAutocompletion: true,
                enableSnippets: true
            }}
            showPrintMargin={false}
            value={code}
            width="100%"
            height="100%"
            style={{ borderRadius: "inherit" }}
        />
    );
}

export default CodeEditor
