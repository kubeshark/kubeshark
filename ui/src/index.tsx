import React from 'react';
import ReactDOM from 'react-dom';
import './index.sass';
import App from './App';
import EntApp from "./EntApp";
import {ToastContainer} from "react-toastify";
import 'react-toastify/dist/ReactToastify.css';

ReactDOM.render(
  <React.StrictMode>
    <>
      {window["isEnt"] ? <EntApp/> : <App/>}
      <ToastContainer
          position="bottom-right"
          autoClose={5000}
          hideProgressBar={false}
          newestOnTop={false}
          closeOnClick
          rtl={false}
          pauseOnFocusLoss
          draggable
          pauseOnHover
      />
    </>
  </React.StrictMode>,
  document.getElementById('root')
);
