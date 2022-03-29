import React, { useEffect, useState } from "react";
import EntryViewer from "./EntryDetailed/EntryViewer";
import { EntryItem } from "./EntryListItem/EntryListItem";
import { makeStyles } from "@material-ui/core";
import Protocol from "../UI/Protocol"
import Queryable from "../UI/Queryable";
import { toast } from "react-toastify";
import { RecoilState, useRecoilState, useRecoilValue } from "recoil";
import focusedEntryIdAtom from "../../recoil/focusedEntryId";
import trafficViewerApi from "../../recoil/TrafficViewerApi";
import TrafficViewerApi from "./TrafficViewerApi";
import TrafficViewerApiAtom from "../../recoil/TrafficViewerApi/atom";
import queryAtom from "../../recoil/query/atom";
import useWindowDimensions, { useRequestTextByWidth } from "../../hooks/WindowDimensionsHook";
import { TOAST_CONTAINER_ID } from "../../configs/Consts";

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
const minSizeDisplayRequestSize = 880;
const EntryTitle: React.FC<any> = ({ protocol, data, elapsedTime }) => {
    const classes = useStyles();
    const request = data.request;
    const response = data.response;

    const { width } = useWindowDimensions();
    const { requestText, responseText, elapsedTimeText } = useRequestTextByWidth(width)

    return <div className={classes.entryTitle}>
        <Protocol protocol={protocol} horizontal={true} />
        {(width > minSizeDisplayRequestSize) && <div style={{ right: "30px", position: "absolute", display: "flex" }}>
            {request && <Queryable
                query={`requestSize == ${data.requestSize}`}
                style={{ margin: "0 18px" }}
                displayIconOnMouseOver={true}
            >
                <div
                    style={{ opacity: 0.5 }}
                    id="entryDetailedTitleRequestSize"
                >
                    {`${requestText}${formatSize(data.requestSize)}`}
                </div>
            </Queryable>}
            {response && <Queryable
                query={`responseSize == ${data.responseSize}`}
                style={{ margin: "0 18px" }}
                displayIconOnMouseOver={true}
            >
                <div
                    style={{ opacity: 0.5 }}
                    id="entryDetailedTitleResponseSize"
                >
                    {`${responseText}${formatSize(data.responseSize)}`}
                </div>
            </Queryable>}
            {response && <Queryable
                query={`elapsedTime >= ${elapsedTime}`}
                style={{ margin: "0 0 0 18px" }}
                displayIconOnMouseOver={true}
            >
                <div
                    style={{ opacity: 0.5 }}
                    id="entryDetailedTitleElapsedTime"
                >
                    {`${elapsedTimeText}${Math.round(elapsedTime)}ms`}
                </div>
            </Queryable>}
        </div>}
    </div>;
};

const EntrySummary: React.FC<any> = ({ entry }) => {
    return <EntryItem
        key={`entry-${entry.id}`}
        entry={entry}
        style={{}}
        headingMode={true}
    />;
};



export const EntryDetailed = () => {

    const focusedEntryId = useRecoilValue(focusedEntryIdAtom);
    const trafficViewerApi = useRecoilValue(TrafficViewerApiAtom as RecoilState<TrafficViewerApi>)
    const query = useRecoilValue(queryAtom);

    const [entryData, setEntryData] = useState(null);

    useEffect(() => {
        if (!focusedEntryId) return;
        setEntryData(null);
        (async () => {
            try {
                const entryData = await trafficViewerApi.getEntry(focusedEntryId, query);
                setEntryData(entryData);
            } catch (error) {
                if (error.response?.data?.type) {
                    toast[error.response.data.type](`Entry[${focusedEntryId}]: ${error.response.data.msg}`, {
                        theme: "colored",
                        autoClose: error.response.data.autoClose,
                        progress: undefined,
                        containerId: TOAST_CONTAINER_ID
                    });
                }
                console.error(error);
            }
        })();
        // eslint-disable-next-line
    }, [focusedEntryId]);

    return <React.Fragment>
        {entryData && <EntryTitle
            protocol={entryData.protocol}
            data={entryData.data}
            elapsedTime={entryData.data.elapsedTime}
        />}
        {entryData && <EntrySummary entry={entryData.base} />}
        <React.Fragment>
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
        </React.Fragment>
    </React.Fragment>
};
