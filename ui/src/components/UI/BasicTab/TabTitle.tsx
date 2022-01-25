import React, { useCallback } from "react"

type Props = {
  title: string
  index: number
  setSelectedTab: (index: number) => void
}

const TabTitle: React.FC<Props> = ({ title, setSelectedTab, index }) => {

  const onClick = useCallback(() => {
    setSelectedTab(index)
  }, [setSelectedTab, index])

  return (
    <li style={{color: "rgb(32,92,245" ,cursor:"pointer"}}>
      <div onClick={onClick}>{title}</div>
      <span style={{    
              cursor: "unset",
              height: "20px",
              margin: "0 20px",
              borderRight: "1px solid #303f9f",
              verticalAlign: "middle"
          }}></span>
    </li>
  )
}

export default TabTitle