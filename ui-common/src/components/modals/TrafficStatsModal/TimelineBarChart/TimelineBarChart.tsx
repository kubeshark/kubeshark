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
    const [commandStats, setCommandStats] = useState(null);
    const [commandNames, setcommandNames] = useState(null);

    useEffect(() => {
        if (!data) return;
        const protocolsBarsData = [];
        const prtcNames = [];
        data.forEach(protocolObj => {
            let newProtocolbj: { [k: string]: any } = {};
            newProtocolbj.timestamp = Utils.getHoursAndMinutes(protocolObj.timestamp);
            protocolObj.protocols.forEach(protocol => {
                newProtocolbj[`${protocol.name}`] = protocol[StatsMode[timeLineBarChartMode]];
                prtcNames.push({ name: protocol.name, color: protocol.color });
            })
            protocolsBarsData.push(newProtocolbj);
        })
        const uniqueObjArray = Utils.creatUniqueObjArrayByProp(prtcNames, "name")
        protocolsBarsData.sort((a, b) => a.timestamp < b.timestamp ? -1 : 1);
        setProtocolStats(protocolsBarsData);
        setProtocolsNamesAndColors(uniqueObjArray);
    }, [data, timeLineBarChartMode])

    useEffect(() => {
        if (selectedProtocol === ALL_PROTOCOLS) {
            setCommandStats(null);
            setcommandNames(null);
            return;
        }
        const commandsNames = [];
        const protocolsCommands = [];
        data.forEach(protocolObj => {
            let newCommandlbj: { [k: string]: any } = {};
            newCommandlbj.timestamp = Utils.getHoursAndMinutes(protocolObj.timestamp);
            protocolObj.protocols.find(protocol => protocol.name === selectedProtocol)?.methods.forEach(command => {
                newCommandlbj[`${command.name}`] = command[StatsMode[timeLineBarChartMode]]
                if (commandsNames.indexOf(command.name) === -1)
                    commandsNames.push(command.name);
            })
            protocolsCommands.push(newCommandlbj);
        })
        protocolsCommands.sort((a, b) => a.timestamp < b.timestamp ? -1 : 1);
        setcommandNames(commandsNames);
        setCommandStats(protocolsCommands);
    }, [data, timeLineBarChartMode, selectedProtocol])

    const bars = useMemo(() => (commandNames || protocolsNamesAndColors).map((entry) => {
        return <Bar key={entry.name || entry} dataKey={entry.name || entry} stackId="a" fill={entry.color || Utils.stringToColor(entry)} />
    }), [protocolsNamesAndColors, commandNames])

    return (
        <div className={styles.barChartContainer}>
            {protocolStats.length > 0 && <BarChart
                width={730}
                height={250}
                data={commandStats || protocolStats}
                margin={{
                    top: 20,
                    right: 30,
                    left: 20,
                    bottom: 5
                }}
            >
                <XAxis dataKey="timestamp" />
                <YAxis tickFormatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value} />
                <Tooltip formatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value + " Requests"} />
                {bars}
            </BarChart>}
        </div>
    );
}
