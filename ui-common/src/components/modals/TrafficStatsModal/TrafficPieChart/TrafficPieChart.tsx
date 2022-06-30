import React, { useEffect, useMemo, useState } from "react";
import { Cell, Legend, Pie, PieChart, Tooltip } from "recharts";
import { Utils } from "../../../../helpers/Utils";
import { ALL_PROTOCOLS, StatsMode as PieChartMode } from "../TrafficStatsModal"

const RADIAN = Math.PI / 180;
const renderCustomizedLabel = ({
  cx,
  cy,
  midAngle,
  innerRadius,
  outerRadius,
  percent,
  index
}: any) => {
  const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
  const x = cx + radius * Math.cos(-midAngle * RADIAN);
  const y = cy + radius * Math.sin(-midAngle * RADIAN);

  if (Number((percent * 100).toFixed(0)) <= 1) return;

  return (
    <text
      x={x}
      y={y}
      fill="white"
      textAnchor={x > cx ? "start" : "end"}
      dominantBaseline="central"
    >
      {`${(percent * 100).toFixed(0)}%`}
    </text>
  );
};

interface TrafficPieChartProps {
  pieChartMode: string;
  data: any;
  selectedProtocol: string;
}

export const TrafficPieChart: React.FC<TrafficPieChartProps> = ({ pieChartMode, data, selectedProtocol }) => {

  const [protocolsStats, setProtocolsStats] = useState([]);
  const [commandStats, setCommandStats] = useState(null);

  useEffect(() => {
    if (!data) return;
    const protocolsPieData = data.map(protocol => {
      return {
        name: protocol.name,
        value: protocol[PieChartMode[pieChartMode]],
        color: protocol.color
      }
    })
    setProtocolsStats(protocolsPieData)
  }, [data, pieChartMode])

  useEffect(() => {
    if (selectedProtocol === ALL_PROTOCOLS) {
      setCommandStats(null);
      return;
    }
    const commandsPieData = data.find(protocol => protocol.name === selectedProtocol)?.methods.map(command => {
      return {
        name: command.name,
        value: command[PieChartMode[pieChartMode]]
      }
    })
    setCommandStats(commandsPieData);
  }, [selectedProtocol, pieChartMode, data])

  const pieLegend = useMemo(() => {
    if (!data) return;
    let legend;
    if (selectedProtocol === ALL_PROTOCOLS) {
      legend = data.map(protocol => <div style={{ marginBottom: 5, display: "flex" }}>
        <div style={{ height: 15, width: 30, background: protocol?.color }} />
        <span style={{ marginLeft: 5 }}>
          {protocol.name}
        </span>
      </div>)
    } else {
      legend = data.find(protocol => protocol.name === selectedProtocol)?.methods.map((method) => <div
        style={{ marginBottom: 5, display: "flex" }}>
        <div style={{ height: 15, width: 30, background: Utils.stringToColor(method.name)}} />
        <span style={{ marginLeft: 5 }}>
          {method.name}
        </span>
      </div>)
    }
    return <div>{legend}</div>;
  }, [data, selectedProtocol])

  return (
    <div>
      {protocolsStats?.length > 0 && <div style={{ width: "100%", display: "flex", justifyContent: "center" }}>
        <PieChart width={300} height={300}>
          <Pie
            data={commandStats || protocolsStats}
            dataKey="value"
            cx={150}
            cy={125}
            labelLine={false}
            label={renderCustomizedLabel}
            outerRadius={125}
            fill="#8884d8">
            {(commandStats || protocolsStats).map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.color || Utils.stringToColor(entry.name)} />)
            )}
          </Pie>
          <Legend wrapperStyle={{ position: "absolute", width: "auto", height: "auto", right: -150, top: 0 }} content={pieLegend} />
          <Tooltip formatter={(value) => pieChartMode === "VOLUME" ? Utils.humanFileSize(value) : value + " Requests"} />
        </PieChart>
      </div>}
    </div>
  );
}
