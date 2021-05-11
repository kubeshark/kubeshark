import React from 'react';
import {HarPage} from "./components/HarPage";
import './App.sass';
import logo from './components/assets/Mizu.svg';

const App = () => {
  return (
      <div className="mizuApp">
        <div className="header">
            <div className="title"><img src={logo} alt="logo"/></div>
            <div className="description">Traffic viewer for Kubernetes</div>
        </div>
        <HarPage/>
      </div>
  );
}

export default App;
