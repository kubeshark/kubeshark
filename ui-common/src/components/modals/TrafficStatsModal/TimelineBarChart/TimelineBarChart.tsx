import styles from "./TimelineBarChart.module.module.sass";
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

const colors = ["#003f5c", "#374c80", "#7a5195", "#bc5090", "#ef5675", "#ff764a", "#ffa600"];

interface TimelineBarChartProps {
    data: any;
}

export const TimelineBarChart: React.FC<TimelineBarChartProps> = ({ data }) => {
    const [protocolStats, setProtocolStats] = useState([]);
    const [protocolsNames, setProtocolsNames] = useState([]);

    const padTo2Digits = (num) => {
        return String(num).padStart(2, '0');
    }

    const getHoursAndMinutes = (protocolTimeKey) => {
        const time = new Date(protocolTimeKey)
        const hoursAndMinutes = padTo2Digits(time.getHours()) + ':' + padTo2Digits(time.getMinutes());
        return hoursAndMinutes;
    }

    useEffect(() => {
        if (!data) return;
        let protocolsBarsData = [];
        let prtcNames = [];
        data.map(protocolObj => {
            let obj: { [k: string]: any } = {};
            obj.time = getHoursAndMinutes(protocolObj.key * 1000);
            protocolObj.protocol.map(protocol => {
                obj[`${protocol.name}`] = protocol.entriesCount;
                prtcNames.push(protocol.name);
            })
            protocolsBarsData.push(obj);
        })
        setProtocolStats(protocolsBarsData);
        setProtocolsNames(prtcNames);
    }, [data])

    const bars = useMemo(() => protocolsNames.map((protoclName, index) => {
        return <Bar key={protoclName} dataKey={protoclName} stackId="a" fill={colors[index]} />
    }), [protocolsNames])

    return (
        <BarChart
            width={500}
            height={300}
            data={protocolStats}
            margin={{
                top: 20,
                right: 30,
                left: 20,
                bottom: 5
            }}
        >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="time" />
            <YAxis />
            <Tooltip />
            <Legend />
            {bars}
        </BarChart>
    );
}
