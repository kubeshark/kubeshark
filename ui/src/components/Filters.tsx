import React, {useRef} from "react";
import styles from './style/Filters.module.sass';
import {Button, Grid} from "@material-ui/core";
import CodeEditor from '@uiw/react-textarea-code-editor';

interface FiltersProps {
    query: string
    setQuery: any
    backgroundColor: string
    ws: any
    openWebSocket: (query: string) => void;
}

export const Filters: React.FC<FiltersProps> = ({query, setQuery, backgroundColor, ws, openWebSocket}) => {
    return <div className={styles.container}>
        <QueryForm
            query={query}
            setQuery={setQuery}
            backgroundColor={backgroundColor}
            ws={ws}
            openWebSocket={openWebSocket}
        />
    </div>;
};

interface QueryFormProps {
    query: string
    setQuery: any
    backgroundColor: string
    ws: any
    openWebSocket: (query: string) => void;
}

export const QueryForm: React.FC<QueryFormProps> = ({query, setQuery, backgroundColor, ws, openWebSocket}) => {

    const formRef = useRef<HTMLFormElement>(null);

    const handleChange = async (e) => {
        setQuery(e.target.value);
    }

    const handleSubmit = (e) => {
        ws.close()
        openWebSocket(query)
        e.preventDefault();
    }

    return <>
        <form
            ref={formRef}
            onSubmit={handleSubmit}
            style={{
                width: '100%',
            }}
        >
            <Grid container spacing={2}>
                <Grid
                    item
                    xs={9}
                    style={{
                        maxHeight: '25vh',
                        overflowY: 'auto',
                    }}
                >
                    <label>
                        <CodeEditor
                            value={query}
                            language="py"
                            placeholder="Mizu Filter Syntax"
                            onChange={handleChange}
                            padding={8}
                            style={{
                                fontSize: 14,
                                backgroundColor: `${backgroundColor}`,
                                fontFamily: 'ui-monospace,SFMono-Regular,SF Mono,Consolas,Liberation Mono,Menlo,monospace',
                            }}
                        />
                    </label>
                </Grid>
                <Grid item xs={3}>
                    <Button type="submit" variant="contained" style={{marginTop: "2px"}}>Apply</Button>
                </Grid>
            </Grid>
        </form>
    </>
}
