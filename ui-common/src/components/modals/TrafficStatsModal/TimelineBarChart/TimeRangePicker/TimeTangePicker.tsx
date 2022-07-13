import React, { useState, Fragment } from 'react';

import {
    EuiSuperDatePicker,
    EuiSpacer,
} from '@elastic/eui';
import dateMath from '@elastic/datemath';

interface TimeRangePickerProps {
    refreshStats: (startTime, endTime) => void;
}

export const TimeRangePicker: React.FC<TimeRangePickerProps> = ({ refreshStats }) => {
    const [recentlyUsedRanges, setRecentlyUsedRanges] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [start, setStart] = useState('now-30m');
    const [end, setEnd] = useState('now');
    const [isPaused, setIsPaused] = useState(true);
    const [refreshInterval, setRefreshInterval] = useState();

    const dateConvertor = (inputStart, inputEnd) => {
        const startMoment = dateMath.parse(inputStart);
        if (!startMoment || !startMoment.isValid()) {
            console.error("Unable to parse start string");
        }
        const endMoment = dateMath.parse(inputEnd, { roundUp: true });
        if (!endMoment || !endMoment.isValid()) {
            console.error("Unable to parse end string");
        }
        return { startMoment: startMoment.format("x"), endMoment: endMoment.format("x") }
    }

    const onTimeChange = ({ start, end }) => {
        const recentlyUsedRange = recentlyUsedRanges.filter(recentlyUsedRange => {
            const isDuplicate =
                recentlyUsedRange.start === start && recentlyUsedRange.end === end;
            return !isDuplicate;
        });
        recentlyUsedRange.unshift({ start, end });
        setStart(start);
        setEnd(end);
        setRecentlyUsedRanges(
            recentlyUsedRange.length > 10
                ? recentlyUsedRange.slice(0, 9)
                : recentlyUsedRange
        );
        const { startMoment, endMoment } = dateConvertor(start, end)
        refreshStats(startMoment, endMoment)
        setIsLoading(true);
        startLoading();
    };

    const onRefresh = ({ start, end, refreshInterval }) => {
        return new Promise(resolve => {
            setTimeout(resolve, 100);
        }).then(() => {
            const { startMoment, endMoment } = dateConvertor(start, end)
            refreshStats(startMoment, endMoment)
        });
    };

    const startLoading = () => {
        setTimeout(stopLoading, 1000);
    };
    const stopLoading = () => {
        setIsLoading(false);
    };

    const onRefreshChange = ({ isPaused, refreshInterval }) => {
        setIsPaused(isPaused);
        setRefreshInterval(refreshInterval);
    };

    return (
        <Fragment>
            <EuiSpacer />
            <EuiSuperDatePicker
                width='auto'
                isLoading={isLoading}
                start={start}
                end={end}
                onTimeChange={onTimeChange}
                onRefresh={onRefresh}
                isPaused={isPaused}
                refreshInterval={refreshInterval}
                onRefreshChange={onRefreshChange}
                recentlyUsedRanges={recentlyUsedRanges}
            />
            <EuiSpacer />
        </Fragment>
    );
};
