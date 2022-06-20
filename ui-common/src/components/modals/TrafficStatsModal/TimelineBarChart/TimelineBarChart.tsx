import styles from "./TimelineBarChart.module.module.sass";
import {StatsMode} from "../TrafficStatsModal"
import React, { useEffect, useMemo, useState } from "react";
import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend
} from "recharts";
import { Utils } from "../../../../helpers/Utils";

interface TimelineBarChartProps {
    timeLineBarChartMode: string;
    data: any;
}

export const TimelineBarChart: React.FC<TimelineBarChartProps> = ({timeLineBarChartMode, data}) => {
    const [protocolStats, setProtocolStats] = useState([]);
    const [protocolsNamesAndColors, setProtocolsNamesAndColors] = useState([]);

    const padTo2Digits = (num) => {
        return String(num).padStart(2, '0');
    }

    const getHoursAndMinutes = (protocolTimeKey) => {
        const time = new Date(protocolTimeKey)
        const hoursAndMinutes = padTo2Digits(time.getHours()) + ':' + padTo2Digits(time.getMinutes());
        return hoursAndMinutes;
    }

    const creatUniqueObjArray = (objArray) => {
        return [
            ...new Map(objArray.map((item) => [item["name"], item])).values(),
        ];
    }

    useEffect(() => {
        if (!data) return;
        let protocolsBarsData = [];
        let prtcNames = [];
        data.map(protocolObj => {
            let obj: { [k: string]: any } = {};
            obj.timestamp = getHoursAndMinutes(protocolObj.timestamp);
            protocolObj.protocols.map(protocol => {
                obj[`${protocol.name}`] = protocol[StatsMode[timeLineBarChartMode]];
                prtcNames.push({name: protocol.name, color: protocol.color});
            })
            protocolsBarsData.push(obj);
        })
        let uniqueObjArray = creatUniqueObjArray(prtcNames);
        setProtocolStats(protocolsBarsData);
        setProtocolsNamesAndColors(uniqueObjArray);
    }, [data,timeLineBarChartMode])

    const bars = useMemo(() => protocolsNamesAndColors.map((protoclToDIsplay) => {
        return <Bar key={protoclToDIsplay.name} dataKey={protoclToDIsplay.name} stackId="a" fill={protoclToDIsplay.color} />
    }), [protocolsNamesAndColors])

    return (
        <div style={{ width: "100%", display: "flex", justifyContent: "center" }}>
            <BarChart
                width={730}
                height={250}
                data={protocolStats}
                margin={{
                    top: 20,
                    right: 30,
                    left: 20,
                    bottom: 5
                }}
            >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" domain={[0,10]}/>
                <YAxis tickFormatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value }/>
                <Tooltip formatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value + " Requests"}/>
                <Legend />
                {bars}
            </BarChart>
        </div>
    );
}
