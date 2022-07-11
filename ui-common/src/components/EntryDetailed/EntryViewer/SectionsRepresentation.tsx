import React from "react";
import { EntryTableSection, EntryBodySection } from "../EntrySections/EntrySections";

enum SectionTypes {
    SectionTable = "table",
    SectionBody = "body",
}

const SectionsRepresentation: React.FC<any> = ({ data, color }) => {
    const sections = []

    if (data) {
        for (const [i, row] of data.entries()) {
            switch (row.type) {
                case SectionTypes.SectionTable:
                    sections.push(
                        <EntryTableSection key={i} title={row.title} color={color} arrayToIterate={JSON.parse(row.data)} />
                    )
                    break;
                case SectionTypes.SectionBody:
                    sections.push(
                        <EntryBodySection key={i} title={row.title} color={color} content={row.data} encoding={row.encoding} contentType={row.mimeType} selector={row.selector} />
                    )
                    break;
                default:
                    break;
            }
        }
    }

    if (sections.length === 0) {
        sections.push(<div>This request or response has no data.</div>);
    }

    return <React.Fragment>{sections}</React.Fragment>;
}

export default SectionsRepresentation
