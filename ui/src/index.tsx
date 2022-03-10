import ReactDOM from 'react-dom';
import './index.sass';
import {ToastContainer} from "react-toastify";
import 'react-toastify/dist/ReactToastify.css';
import {RecoilRoot} from "recoil";
import App from './App';



ReactDOM.render( <>
    <RecoilRoot>
        <App/>
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
    </RecoilRoot>
</>,
document.getElementById('root'));
