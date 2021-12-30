import React from 'react';
import ReactDOM from 'react-dom';
import './index.sass';
import App from './App';
import EntApp from "./EntApp";

ReactDOM.render(
  <React.StrictMode>
      {window["isEnt"] ? <EntApp/> : <App/>}
  </React.StrictMode>,
  document.getElementById('root')
);
