import ReactDOM from 'react-dom';
import './index.sass';
import {ToastContainer} from "react-toastify";
import 'react-toastify/dist/ReactToastify.min.css';
import {RecoilRoot} from "recoil";
import App from './App';
import { TOAST_CONTAINER_ID } from './consts';

ReactDOM.render( <>
    <RecoilRoot>
        <App/>
        <ToastContainer enableMultiContainer containerId={TOAST_CONTAINER_ID} 
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
    </RecoilRoot>
</>,
document.getElementById('root'));
