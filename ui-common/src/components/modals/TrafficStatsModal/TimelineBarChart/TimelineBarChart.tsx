import styles from "./TimelineBarChart.module.sass";
import { ALL_PROTOCOLS, StatsMode } from "../TrafficStatsModal"
import React, { useEffect, useMemo, useState } from "react";
import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    Tooltip,
} from "recharts";
import { Utils } from "../../../../helpers/Utils";

interface TimelineBarChartProps {
    timeLineBarChartMode: string;
    data: any;
    selectedProtocol: string;
}

export const TimelineBarChart: React.FC<TimelineBarChartProps> = ({ timeLineBarChartMode, data, selectedProtocol }) => {
    const [protocolStats, setProtocolStats] = useState([]);
    const [protocolsNamesAndColors, setProtocolsNamesAndColors] = useState([]);
    const [methodsStats, setMethodsStats] = useState(null);
    const [methodsNamesAndColors, setMethodsNamesAndColors] = useState(null);
    
    useEffect(() => {
        if (!data) return;
        const protocolsBarsData = [];
        const prtcNames = [];
        data.sort((a, b) => a.timestamp < b.timestamp ? -1 : 1).forEach(protocolObj => {
            let newProtocolObj: { [k: string]: any } = {};
            newProtocolObj.timestamp = Utils.getHoursAndMinutes(protocolObj.timestamp);
            protocolObj.protocols.forEach(protocol => {
                newProtocolObj[`${protocol.name}`] = protocol[StatsMode[timeLineBarChartMode]];
                prtcNames.push({ name: protocol.name, color: protocol.color });
            })
            protocolsBarsData.push(newProtocolObj);
        })
        const uniqueObjArray = Utils.creatUniqueObjArrayByProp(prtcNames, "name")
        setProtocolStats(protocolsBarsData);
        setProtocolsNamesAndColors(uniqueObjArray);
    }, [data, timeLineBarChartMode])

    useEffect(() => {
        if (selectedProtocol === ALL_PROTOCOLS) {
            setMethodsStats(null);
            setMethodsNamesAndColors(null);
            return;
        }
        const protocolsMethodsNamesAndColors = [];
        const protocolsMethods = [];
        data.sort((a, b) => a.timestamp < b.timestamp ? -1 : 1).forEach(protocolObj => {
            let newMethodObj: { [k: string]: any } = {};
            newMethodObj.timestamp = Utils.getHoursAndMinutes(protocolObj.timestamp);
            protocolObj.protocols.find(protocol => protocol.name === selectedProtocol)?.methods.forEach(method => {
                newMethodObj[`${method.name}`] = method[StatsMode[timeLineBarChartMode]]
                protocolsMethodsNamesAndColors.push({name: method.name, color: method.color});
            })
            protocolsMethods.push(newMethodObj);
        })
        const uniqueObjArray = Utils.creatUniqueObjArrayByProp(protocolsMethodsNamesAndColors, "name")
        setMethodsNamesAndColors(uniqueObjArray);
        setMethodsStats(protocolsMethods);
    }, [data, timeLineBarChartMode, selectedProtocol])

    const bars = useMemo(() => (methodsNamesAndColors || protocolsNamesAndColors).map((entry) => {
        return <Bar key={entry.name} dataKey={entry.name} stackId="a" fill={entry.color} />
    }), [protocolsNamesAndColors, methodsNamesAndColors])

    const renderTick = (tickProps) => {
        const { x, y, payload } = tickProps;
        const { index, value } = payload;

        if (protocolStats.length > 5) {
            if (index % 3 === 0) {
                return <text x={x} y={y + 10} textAnchor="end">{`${value}`}</text>;
            }
            return null;
        }
        else {
            return <text x={x} y={y + 10} textAnchor="end">{`${value}`}</text>;
        }
    };

    return (
        <div className={styles.barChartContainer}>
            {protocolStats.length > 0 && <BarChart
                width={750}
                height={250}
                data={methodsStats || protocolStats}
                barCategoryGap={1}
                margin={{
                    top: 20,
                    right: 30,
                    left: 20,
                    bottom: 5
                }}
            >
                <XAxis dataKey="timestamp" tick={renderTick} tickLine={false} interval="preserveStart" />
                <YAxis tickFormatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value} interval="preserveEnd"/>
                <Tooltip formatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value + " Requests"} />
                {bars}
            </BarChart>}
        </div>
    );
}
