import React, {useState} from 'react';
import styles from './EntryViewer.module.sass';
import Tabs from "../../UI/Tabs/Tabs";
import {EntryTableSection, EntryBodySection} from "../EntrySections/EntrySections";

enum SectionTypes {
    SectionTable = "table",
    SectionBody = "body",
}

const SectionsRepresentation: React.FC<any> = ({data, color}) => {
    const sections = []

    if (data) {
        for (const [i, row] of data.entries()) {
            switch (row.type) {
                case SectionTypes.SectionTable:
                    sections.push(
                        <EntryTableSection key={i} title={row.title} color={color} arrayToIterate={JSON.parse(row.data)}/>
                    )
                    break;
                case SectionTypes.SectionBody:
                    sections.push(
                        <EntryBodySection key={i} title={row.title} color={color} content={row.data} encoding={row.encoding} contentType={row.mimeType} selector={row.selector}/>
                    )
                    break;
                default:
                    break;
            }
        }
    }

    return <React.Fragment>{sections}</React.Fragment>;
}

const AutoRepresentation: React.FC<any> = ({representation, elapsedTime, color}) => {
    const TABS = [
        {
            tab: 'Request'
        }
    ];
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);

    // Don't fail even if `representation` is an empty string
    if (!representation) {
        return <React.Fragment></React.Fragment>;
    }

    const {request, response} = JSON.parse(representation);

    let responseTabIndex = 0;

    if (response) {
        TABS.push(
            {
                tab: 'Response',
            }
        );
        responseTabIndex = TABS.length - 1;
    }

    return <div className={styles.Entry}>
        {<div className={styles.body}>
            <div className={styles.bodyHeader}>
                <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned/>
            </div>
            {currentTab === TABS[0].tab && <React.Fragment>
                <SectionsRepresentation data={request} color={color}/>
            </React.Fragment>}
            {response && currentTab === TABS[responseTabIndex].tab && <React.Fragment>
                <SectionsRepresentation data={response} color={color}/>
            </React.Fragment>}
        </div>}
    </div>;
}

interface Props {
    representation: any;
    color: string;
    elapsedTime: number;
}

const EntryViewer: React.FC<Props> = ({representation, elapsedTime, color}) => {
    return <AutoRepresentation
        representation={representation}
        elapsedTime={elapsedTime}
        color={color}
    />
};

export default EntryViewer;
