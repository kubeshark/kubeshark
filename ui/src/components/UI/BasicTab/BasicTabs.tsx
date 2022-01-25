import React, { ReactElement, useState } from "react";
import TabTitle from "./TabTitle";
import styles from '../style/BasicTabs.module.sass'


type Props = {
  children: ReactElement[]
}

const Tabs: React.FC<Props> = ({ children }) => {
  const [selectedTab, setSelectedTab] = useState(0);

  return (
    <div>
      <ul style={{"display":"flex","justifyContent":"space-between", listStyleType:"none"}}>
        {children.map((item, index) => {
          <TabTitle
            key={index}
            title={item.props.title}
            index={index}
            setSelectedTab={setSelectedTab}
          />
          {<span style={{    
            cursor: "unset",
            height: "20px",
            margin: "0 20px",
            borderRight: "1px solid #303f9f",
            verticalAlign: "middle"
            }}></span>}
    })}
      </ul>
      {children[selectedTab]}
    </div>
  )
}

export default Tabs