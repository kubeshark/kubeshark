import React, {useEffect} from "react";
import styles from './style/HarFilters.module.sass';
import {HARFilterSelect} from "./HARFilterSelect";
import {TextField} from "@material-ui/core";

export const HarFilters: React.FC = () => {

    return <div className={styles.container}>
        <ServiceFilter/>
        <MethodFilter/>
        <StatusTypesFilter/>
        <SourcesFilter/>
        <FetchModeFilter/>
        <PathFilter/>
    </div>;
};

const _toUpperCase = v => v.toUpperCase();

const FilterContainer: React.FC = ({children}) => {
    return <div className={styles.filterContainer}>
        {children}
    </div>;
};

const ServiceFilter: React.FC = () => {
    const providerIds = []; //todo
    const selectedServices = []; //todo

    return <FilterContainer>
        <HARFilterSelect
            items={providerIds}
            value={selectedServices}
            onChange={(val) => {
                //todo: harStore.updateFilter({toggleService: val})
            }}
            allowMultiple={true}
            label={"Services"}
            transformDisplay={_toUpperCase}
        />
    </FilterContainer>

};

const BROWSER_SOURCE = "_BROWSER_";

const SourcesFilter: React.FC = () => {

    const sources = []; //todo
    const selectedSource = null; //todo

    useEffect(() => {
        //todo: fetch sources
    }, []);

    return <FilterContainer>
        <HARFilterSelect
            items={sources}
            value={selectedSource}
            onChange={(val) => {
                //todo: harStore.updateFilter({toggleSource: val});
            }}
            allowMultiple={true}
            label={"Sources"}
            transformDisplay={item => item === BROWSER_SOURCE ? "BROWSER" : item.toUpperCase()}
        />
    </FilterContainer>

};

enum HARFetchMode {
    UP_TO_REVISION = "Up to revision",
    ALL = "All",
    QUEUED = "Unprocessed"
}

const FetchModeFilter: React.FC = () => {

    const selectedHarFetchMode = null;

    return <FilterContainer>
        <HARFilterSelect
            items={Object.values(HARFetchMode)}
            value={selectedHarFetchMode}
            onChange={(val) => {
                // selectedModelStore.har.setHarFetchMode(val);
                // selectedModelStore.har.data.reset();
                // selectedModelStore.har.data.fetch();
                //todo
            }}
            label={"Processed"}
        />
    </FilterContainer>

};

enum HTTPMethod {
    GET = "get",
    PUT = "put",
    POST = "post",
    DELETE = "delete",
    OPTIONS="options",
    PATCH = "patch"
}

const MethodFilter: React.FC = () => {

    const selectedMethods = [];

    return <FilterContainer>
        <HARFilterSelect
            items={Object.values(HTTPMethod)}
            allowMultiple={true}
            value={selectedMethods}
            onChange={(val) => {
                // harStore.updateFilter({toggleMethod: val}) todo
            }}
            transformDisplay={_toUpperCase}
            label={"Methods"}
        />
    </FilterContainer>;
};

enum StatusType {
    SUCCESS = "success",
    ERROR = "error"
}

const StatusTypesFilter: React.FC = () => {

    const selectedStatusTypes = [];

    return <FilterContainer>
        <HARFilterSelect
            items={Object.values(StatusType)}
            allowMultiple={true}
            value={selectedStatusTypes}
            onChange={(val) => {
                // harStore.updateFilter({toggleStatusType: val}) todo
            }}
            transformDisplay={_toUpperCase}
            label="Status"
        />
    </FilterContainer>;
};

// TODO path search is inclusive of the qs -> we want to avoid this - TRA-1681
const PathFilter: React.FC = () => {

    const onFilterChange = (value) => {
        // harStore.updateFilter({setPathSearch: value}); todo
    }

    return <FilterContainer>
        <div className={styles.filterLabel}>Path</div>
        <div>
            <TextField variant="outlined" className={styles.filterText} style={{minWidth: '150px'}} onKeyDown={(e: any) => e.key === "Enter" && onFilterChange(e.target.value)}/>
        </div>
    </FilterContainer>;
};

