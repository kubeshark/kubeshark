import * as React from 'react';

export type EntryDetailedConfig = {
    isReplayEnabled: boolean
}

export interface IEntryDetailedProps {
    config: EntryDetailedConfig
}

export type TEntryDetailedContext = IEntryDetailedProps

export const EntryDetailedContext = React.createContext<TEntryDetailedContext | null>(null);

const EntryDetailedProvider: React.FC<IEntryDetailedProps> = ({ config, children }) => {
    return <EntryDetailedContext.Provider value={{ config }}>{children}</EntryDetailedContext.Provider>;
};

export default EntryDetailedProvider;
