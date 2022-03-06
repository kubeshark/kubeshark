import React, {useEffect, useState} from "react";
import EntryViewer from "./EntryDetailed/EntryViewer";
import {EntryItem} from "./EntryListItem/EntryListItem";
import {makeStyles} from "@material-ui/core";
import Protocol from "./UI/Protocol"
import Queryable from "./UI/Queryable";
import {toast} from "react-toastify";
import {useRecoilValue} from "recoil";
import focusedEntryIdAtom from "../recoil/focusedEntryId";
import Api from "../helpers/api";
import queryAtom from "../recoil/query";

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

export const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;

const EntryTitle: React.FC<any> = ({protocol, data, bodySize, elapsedTime}) => {
    const classes = useStyles();
    const response = data.response;

    return <div className={classes.entryTitle}>
        <Protocol protocol={protocol} horizontal={true}/>
        <div style={{right: "30px", position: "absolute", display: "flex"}}>
            {response && <Queryable
                query={`response.bodySize == ${bodySize}`}
                style={{margin: "0 18px"}}
                displayIconOnMouseOver={true}
            >
                <div
                    style={{opacity: 0.5}}
                    id="entryDetailedTitleBodySize"
                >
                    {formatSize(bodySize)}
                </div>
            </Queryable>}
            {response && <Queryable
                query={`elapsedTime >= ${elapsedTime}`}
                style={{marginRight: 18}}
                displayIconOnMouseOver={true}
            >
                <div
                    style={{opacity: 0.5}}
                    id="entryDetailedTitleElapsedTime"
                >
                    {Math.round(elapsedTime)}ms
                </div>
            </Queryable>}
        </div>
    </div>;
};

const EntrySummary: React.FC<any> = ({entry}) => {
    return <EntryItem
        key={`entry-${entry.id}`}
        entry={entry}
        style={{}}
        headingMode={true}
    />;
};

const api = Api.getInstance();

export const EntryDetailed = () => {

    const focusedEntryId = useRecoilValue(focusedEntryIdAtom);
    const query = useRecoilValue(queryAtom);
    const [entryData, setEntryData] = useState(null);

    useEffect(() => {
        if (!focusedEntryId) return;
        setEntryData(null);
        (async () => {
            try {
                const entryData = await api.getEntry(focusedEntryId, query);
                setEntryData(entryData);
            } catch (error) {
                if (error.response?.data?.type) {
                    toast[error.response.data.type](`Entry[${focusedEntryId}]: ${error.response.data.msg}`, {
                        position: "bottom-right",
                        theme: "colored",
                        autoClose: error.response.data.autoClose,
                        hideProgressBar: false,
                        closeOnClick: true,
                        pauseOnHover: true,
                        draggable: true,
                        progress: undefined,
                    });
                }
                console.error(error);
            }
        })();
        // eslint-disable-next-line
    }, [focusedEntryId]);

    return <>
        {entryData && <EntryTitle
            protocol={entryData.protocol}
            data={entryData.data}
            bodySize={entryData.bodySize}
            elapsedTime={entryData.data.elapsedTime}
        />}
        {entryData && <EntrySummary entry={entryData.data}/>}
        <>
            {entryData && <EntryViewer
                representation={entryData.representation}
                isRulesEnabled={entryData.isRulesEnabled}
                rulesMatched={entryData.rulesMatched}
                contractStatus={entryData.data.contractStatus}
                requestReason={entryData.data.contractRequestReason}
                responseReason={entryData.data.contractResponseReason}
                contractContent={entryData.data.contractContent}
                elapsedTime={entryData.data.elapsedTime}
                color={entryData.protocol.backgroundColor}
            />}
        </>
    </>
};
