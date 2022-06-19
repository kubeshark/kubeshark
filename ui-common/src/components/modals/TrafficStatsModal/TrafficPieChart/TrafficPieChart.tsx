import React, {useEffect, useMemo, useState} from "react";
import styles from "./TrafficPieChart.module.sass";
import {Cell, Legend, Pie, PieChart, Tooltip} from "recharts";
import {Utils} from "../../../../helpers/Utils";

enum PieChartMode {
  REQUESTS = "entriesCount",
  VOLUME = "volumeSizeBytes"
}

const COLORS = ['#e6194b', '#3cb44b', '#ffe119', '#4363d8', '#f58231', '#911eb4', '#46f0f0', '#f032e6', '#bcf60c', '#fabebe', '#008080', '#e6beff', '#9a6324', '#fffac8', '#800000', '#aaffc3', '#808000', '#ffd8b1', '#000075', '#808080', '#ffffff', '#000000'];

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
}

export const TrafficPieChart: React.FC<TrafficPieChartProps> = ({pieChartMode , data}) => {

  const [protocolsStats, setProtocolsStats] = useState([]);
  const [commandStats, setCommandStats] = useState(null);
  const [selectedProtocol, setSelectedProtocol] = useState(null as string);

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
    if (!selectedProtocol) {
      setCommandStats(null);
      return;
    }
    const commandsPieData = data.find(protocol => protocol.name === selectedProtocol).methods.map(command => {
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
    if (!selectedProtocol) {
      legend = data.map(protocol => <div style={{marginBottom: 5, display: "flex"}}>
        <div style={{height: 15, width: 30, background: protocol?.color}}/>
        <span style={{marginLeft: 5}}>
          {protocol.name}
        </span>
      </div>)
    } else {
      legend = data.find(protocol => protocol.name === selectedProtocol).methods.map((method, index) => <div
        style={{marginBottom: 5, display: "flex"}}>
        <div style={{height: 15, width: 30, background: COLORS[index % COLORS.length]}}/>
        <span style={{marginLeft: 5}}>
          {method.name}
        </span>
      </div>)
    }
    return <div>{legend}</div>;
  }, [data, selectedProtocol])

  return (
    <div>
      <div className={styles.breadCrumbsContainer}>
        {selectedProtocol && <div className={styles.breadCrumbs}>
          <span className={styles.clickableTag} onClick={() => setSelectedProtocol(null)}>protocols</span>
          <span>/</span>
          <span className={styles.nonClickableTag}>{selectedProtocol}</span>
        </div>}
      </div>

      {protocolsStats?.length > 0 && <div style={{width: "100%", display: "flex", justifyContent: "center"}}>
        <PieChart width={300} height={300}>
          <Pie
            data={commandStats || protocolsStats}
            dataKey="value"
            cx={150}
            cy={125}
            labelLine={false}
            label={renderCustomizedLabel}
            outerRadius={125}
            fill="#8884d8"
            onClick={(section) => !commandStats && setSelectedProtocol(section.name)}>
            {(commandStats || protocolsStats).map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.color || COLORS[index % COLORS.length]}/>)
            )}
          </Pie>
          <Legend wrapperStyle={{position: "absolute", width: "auto", height: "auto", right: -150, top: 0}} content={pieLegend}/>
          <Tooltip formatter={(value) => pieChartMode === "VOLUME" ? Utils.humanFileSize(value) : value + " Requests"}/>
        </PieChart>
      </div>}
    </div>
  );
}
