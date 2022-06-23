import React, { } from "react";
import AceEditor from "react-ace";
import ReactAce from "react-ace/lib/ace";
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
    name?: string,
    onChange?: (code: string) => void,
    isDisabled?: boolean,
    className?: string,
    variableHeight?: boolean,
    language?: string,
    errorMessage?: string,
    hideTooltip?: boolean,
    hideGutter?: boolean
}

export const CodeEditor = React.forwardRef<ReactAce, CodeEditorProps>((
    {
        code,
        onChange,
        isDisabled = false,
        className,
        name,
        variableHeight = false,
        language = 'python',
        errorMessage,
        hideTooltip = false,
        hideGutter = false
    }, ref) => {
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
        />

    );
});
