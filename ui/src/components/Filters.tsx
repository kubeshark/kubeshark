import React from "react";
import styles from './style/Filters.module.sass';
import {FilterSelect} from "./UI/FilterSelect";
import {TextField, Button} from "@material-ui/core";
import {ALL_KEY} from "./UI/Select";

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

class QueryForm extends React.Component<QueryFormProps, { value: string }> {
    constructor(props) {
        super(props);
        this.state = {value: ''};

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleChange(event) {
        this.setState({value: event.target.value});
    }

    handleSubmit(event) {
        this.props.ws.close()
        this.props.openWebSocket(this.state.value)
        // alert('A name was submitted: ' + this.state.value);
        event.preventDefault();
    }

    render() {
    return (
        <form onSubmit={this.handleSubmit}>
        <label>
            <TextField value={this.state.value} onChange={this.handleChange} variant="outlined" className={styles.filterText} style={{minWidth: '450px'}} placeholder="Mizu Filter Syntax"/>
        </label>
        <Button type="submit" variant="contained" style={{marginLeft: "10px"}}>Apply</Button>
        </form>
    );
    }
}
