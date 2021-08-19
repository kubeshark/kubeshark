import React from "react";
import styles from './style/Filters.module.sass';
import {FilterSelect} from "./UI/FilterSelect";
import {TextField} from "@material-ui/core";
import {ALL_KEY} from "./UI/Select";

interface HarFiltersProps {
    methodsFilter: Array<string>;
    setMethodsFilter: (methods: Array<string>) => void;
    statusFilter: Array<string>;
    setStatusFilter: (methods: Array<string>) => void;
    pathFilter: string
    setPathFilter: (val: string) => void;
}

export const Filters: React.FC<HarFiltersProps> = ({methodsFilter, setMethodsFilter, statusFilter, setStatusFilter, pathFilter, setPathFilter}) => {

    return <div className={styles.container}>
        <MethodFilter methodsFilter={methodsFilter} setMethodsFilter={setMethodsFilter}/>
        <StatusTypesFilter statusFilter={statusFilter} setStatusFilter={setStatusFilter}/>
        <PathFilter pathFilter={pathFilter} setPathFilter={setPathFilter}/>
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

interface PathFilterProps {
    pathFilter: string;
    setPathFilter: (val: string) => void;
}

const PathFilter: React.FC<PathFilterProps> = ({pathFilter, setPathFilter}) => {

    return <FilterContainer>
        <div className={styles.filterLabel}>Path</div>
        <div>
            <TextField value={pathFilter} variant="outlined" className={styles.filterText} style={{minWidth: '150px'}} onChange={(e: any) => setPathFilter(e.target.value)}/>
        </div>
    </FilterContainer>;
};

