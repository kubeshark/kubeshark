import React from 'react';
import {HarPage} from "./components/HarPage";

const App = () => {
  return (
      <div style={{backgroundColor: "#090b14", width: "100%"}}>
        <div style={{height: 100, display: "flex", alignItems: "center", paddingLeft: 24}}>
            <div style={{fontSize: 45, letterSpacing: 2}}>MIZU</div>
            <div style={{marginLeft: 10, color: "rgba(255,255,255,0.5)", paddingTop: 15, fontSize: 16}}>Traffic viewer for Kubernetes</div>
        </div>
        <HarPage/>
      </div>
  );
}

export default App;
