import React from "react";
import EntryViewer from "./EntryDetailed/EntryViewer";
import {EntryItem} from "./EntryListItem/EntryListItem";
import {makeStyles} from "@material-ui/core";
import Protocol from "./UI/Protocol"
import Queryable from "./UI/Queryable";

const useStyles = makeStyles(() => ({
    entryTitle: {
        display: 'flex',
        minHeight: 20,
        maxHeight: 46,
        alignItems: 'center',
        marginBottom: 4,
        marginLeft: 6,
        padding: 2,
        paddingBottom: 0
    },
    entrySummary: {
        display: 'flex',
        minHeight: 36,
        maxHeight: 46,
        alignItems: 'center',
        marginBottom: 4,
        padding: 5,
        paddingBottom: 0
    }
}));

interface EntryDetailedProps {
    entryData: any
    updateQuery: any
}

export const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;

const EntryTitle: React.FC<any> = ({protocol, data, bodySize, elapsedTime, updateQuery}) => {
    const classes = useStyles();
    const response = data.response;

    return <div className={classes.entryTitle}>
        <Protocol protocol={protocol} horizontal={true} updateQuery={updateQuery}/>
        <div style={{right: "30px", position: "absolute", display: "flex"}}>
            {response && <Queryable
                query={`response.bodySize == ${bodySize}`}
                updateQuery={updateQuery}
                style={{margin: "0 18px"}}
                displayIconOnMouseOver={true}
            >
                <div
                    style={{opacity: 0.5}}
                >
                    {formatSize(bodySize)}
                </div>
            </Queryable>}
            {response && <Queryable
                query={`elapsedTime >= ${elapsedTime}`}
                updateQuery={updateQuery}
                style={{marginRight: 18}}
                displayIconOnMouseOver={true}
            >
                <div
                    style={{opacity: 0.5}}
                >
                    {Math.round(elapsedTime)}ms
                </div>
            </Queryable>}
        </div>
    </div>;
};

const EntrySummary: React.FC<any> = ({data, updateQuery}) => {
    const entry = data.base;

    return <EntryItem
        key={`entry-${entry.id}`}
        entry={entry}
        style={{}}
        updateQuery={updateQuery}
        headingMode={true}
    />;
};

export const EntryDetailed: React.FC<EntryDetailedProps> = ({entryData, updateQuery}) => {
    return <>
        <EntryTitle
            protocol={entryData.protocol}
            data={entryData.data}
            bodySize={entryData.bodySize}
            elapsedTime={entryData.data.elapsedTime}
            updateQuery={updateQuery}
        />
        {entryData.data && <EntrySummary data={entryData.data} updateQuery={updateQuery}/>}
        <>
            {entryData.data && <EntryViewer
                representation={entryData.representation}
                isRulesEnabled={entryData.isRulesEnabled}
                rulesMatched={entryData.rulesMatched}
                contractStatus={entryData.data.contractStatus}
                requestReason={entryData.data.contractRequestReason}
                responseReason={entryData.data.contractResponseReason}
                contractContent={entryData.data.contractContent}
                elapsedTime={entryData.data.elapsedTime}
                color={entryData.protocol.backgroundColor}
                updateQuery={updateQuery}
            />}
        </>
    </>
};
