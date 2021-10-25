import React, {useRef, useState} from "react";
import styles from './style/Filters.module.sass';
import {Button} from "@material-ui/core";
import CodeEditor from '@uiw/react-textarea-code-editor';

interface FiltersProps {
    ws: any
    openWebSocket: (query: string) => void;
}

export const Filters: React.FC<FiltersProps> = ({ws, openWebSocket}) => {
    return <div className={styles.container}>
        <QueryForm ws={ws} openWebSocket={openWebSocket}/>
    </div>;
};

interface QueryFormProps {
    ws: any
    openWebSocket: (query: string) => void;
}

export const QueryForm: React.FC<QueryFormProps> = ({ws, openWebSocket}) => {

    const [value, setValue] = useState("");
    const formRef = useRef<HTMLFormElement>(null);

    const handleChange = (e) => {
        setValue(e.target.value);
    }

    const handleSubmit = (e) => {
        ws.close()
        openWebSocket(value)
        e.preventDefault();
    }

    return <>
        <form ref={formRef} onSubmit={handleSubmit}>
        <label>
            <CodeEditor
                value={value}
                language="py"
                placeholder="Mizu Filter Syntax"
                onChange={handleChange}
                padding={8}
                style={{
                    fontSize: 14,
                    backgroundColor: "#f5f5f5",
                    fontFamily: 'ui-monospace,SFMono-Regular,SF Mono,Consolas,Liberation Mono,Menlo,monospace',
                    display: 'inline-flex',
                    minWidth: '450px',
                }}
            />
        </label>
        <Button type="submit" variant="contained" style={{marginLeft: "10px"}}>Apply</Button>
        </form>
    </>
}
