import styles from "./TimelineBarChart.module.sass";
import React, { useEffect, useMemo, useState } from "react";
import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    Tooltip,
} from "recharts";
import { Utils } from "../../../../helpers/Utils";
import { ALL_PROTOCOLS, StatsMode } from "../consts";

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
            newProtocolObj.timestamp = Utils.formatDate(protocolObj.timestamp);
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
            newMethodObj.timestamp = Utils.formatDate(protocolObj.timestamp);
            protocolObj.protocols.find(protocol => protocol.name === selectedProtocol)?.methods.forEach(method => {
                newMethodObj[`${method.name}`] = method[StatsMode[timeLineBarChartMode]]
                protocolsMethodsNamesAndColors.push({ name: method.name, color: method.color });
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
        const { offset, value } = payload;
        const pathX = Math.floor(x - offset) + 0.5;

        return <React.Fragment>
            <text x={pathX} y={y + 10} textAnchor="middle" className={styles.axisText}>{`${value}`}</text>;
            <path d={`M${pathX},${y - 4}v${-10}`} stroke="red" />;
        </React.Fragment>
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
                <XAxis dataKey="timestamp" tickLine={false} tick={renderTick} interval="preserveStart"/>
                <YAxis tickFormatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value} interval="preserveEnd" />
                <Tooltip formatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value + " Requests"} />
                {bars}
            </BarChart>}
        </div>
    );
}
