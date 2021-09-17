import React, {useRef, useState} from "react";
import styles from './style/Filters.module.sass';
import {FilterSelect} from "./UI/FilterSelect";
import {Button} from "@material-ui/core";
import {ALL_KEY} from "./UI/Select";
import CodeEditor from '@uiw/react-textarea-code-editor';

interface FiltersProps {
    methodsFilter: Array<string>;
    setMethodsFilter: (methods: Array<string>) => void;
    statusFilter: Array<string>;
    setStatusFilter: (methods: Array<string>) => void;
    pathFilter: string
    setPathFilter: (val: string) => void;
    ws: any
    openWebSocket: (query: string) => void;
}

export const Filters: React.FC<FiltersProps> = ({methodsFilter, setMethodsFilter, statusFilter, setStatusFilter, pathFilter, setPathFilter, ws, openWebSocket}) => {

    return <div className={styles.container}>
        <QueryForm ws={ws} openWebSocket={openWebSocket}/>
    </div>;
};

const _toUpperCase = v => v.toUpperCase();

const FilterContainer: React.FC = ({children}) => {
    return <div className={styles.filterContainer}>
        {children}
    </div>;
};

enum HTTPMethod {
    GET = "get",
    PUT = "put",
    POST = "post",
    DELETE = "delete",
    OPTIONS="options",
    PATCH = "patch"
}

interface MethodFilterProps {
    methodsFilter: Array<string>;
    setMethodsFilter: (methods: Array<string>) => void;
}

const MethodFilter: React.FC<MethodFilterProps> = ({methodsFilter, setMethodsFilter}) => {

    const methodClicked = (val) => {
        if(val === ALL_KEY) {
            setMethodsFilter([]);
            return;
        }
        if(methodsFilter.includes(val)) {
            setMethodsFilter(methodsFilter.filter(method => method !== val))
        } else {
            setMethodsFilter([...methodsFilter, val]);
        }
    }

    return <FilterContainer>
        <FilterSelect
            items={Object.values(HTTPMethod)}
            allowMultiple={true}
            value={methodsFilter}
            onChange={(val) => methodClicked(val)}
            transformDisplay={_toUpperCase}
            label={"Methods"}
        />
    </FilterContainer>;
};

export enum StatusType {
    SUCCESS = "success",
    ERROR = "error"
}

interface StatusTypesFilterProps {
    statusFilter: Array<string>;
    setStatusFilter: (methods: Array<string>) => void;
}

const StatusTypesFilter: React.FC<StatusTypesFilterProps> = ({statusFilter, setStatusFilter}) => {

    const statusClicked = (val) => {
        if(val === ALL_KEY) {
            setStatusFilter([]);
            return;
        }
        setStatusFilter([val]);
    }

    return <FilterContainer>
        <FilterSelect
            items={Object.values(StatusType)}
            allowMultiple={true}
            value={statusFilter}
            onChange={(val) => statusClicked(val)}
            transformDisplay={_toUpperCase}
            label="Status"
        />
    </FilterContainer>;
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
