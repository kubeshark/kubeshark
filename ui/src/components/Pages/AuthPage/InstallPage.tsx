import { Button } from "@material-ui/core";
import React, { useState,useRef } from "react";
import { adminUsername } from "../../../consts";
import Api, { FormValidationErrorType } from "../../../helpers/api";
import { toast } from 'react-toastify';
import LoadingOverlay from "../../LoadingOverlay";
import { useCommonStyles } from "../../../helpers/commonStyle";
import useKeyPress from "../../../hooks/useKeyPress"
import shortcutsKeyboard from "../../../configs/shortcutsKeyboard"
import { useNavigate, useParams } from "react-router-dom";



const api = Api.getInstance();

interface InstallPageProps {
    onFirstLogin: () => void;
}

export const InstallPage: React.FC<InstallPageProps> = ({onFirstLogin}) => {

    const formRef = useRef(null);
    const classes = useCommonStyles();
    const [isLoading, setIsLoading] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordConfirm, setPasswordConfirm] = useState("");
    const {inviteToken} = useParams()
    const navigate = useNavigate();
    
    const onFormSubmit = async () => {
        if (password.length < 4) {
            toast.error("Password must be at least 4 characters long");
            return;
        } else if (password !== passwordConfirm) {
            toast.error("Passwords do not match");
            return;
        }

        if(inviteToken){
            registerUser(async ()=> await api.recoverUser({password: password, inviteToken: inviteToken}))
        }
        else{
            registerUser(async () => await api.setupAdminUser(adminUsername, password))
        }
    }

    const registerUser = async(registerFunc) => {
        try {
    
            setIsLoading(true);           
            await registerFunc();
            if (!await api.isAuthenticationNeeded()) {
                navigate('/');
                onFirstLogin();
            }
        } catch (e) {
            if (e.type === FormValidationErrorType) {
                for (const messages of Object.values(e.messages) as any[]) {
                    for (const message of messages) {
                        toast.error(message.text);
                    }
                }
            } else {
                toast.error("An unknown error has occured");
            }
            console.error(e);
        } finally {
            setIsLoading(false);
        }
    }

    useKeyPress(shortcutsKeyboard.enter, onFormSubmit, formRef.current);

    return <div className="centeredForm" ref={formRef}>
            {isLoading && <LoadingOverlay/>}
            <div className="form-title left-text">Setup</div>
            {inviteToken && <span className="form-subtitle">You have been invited to join Mizu, please set password in order to complete the registration process</span>}
            {!inviteToken && <span className="form-subtitle">Welcome to Mizu, please set up the admin user to continue</span>}
            {!inviteToken && <div className="form-input">
                <label htmlFor="inputUsername">Username</label>
                <input id="inputUsername" className={classes.textField} value={adminUsername} disabled={true} />
            </div>}
            <div className="form-input">
                <label htmlFor="inputUsername">Password</label>
                <input id="inputUsername" className={classes.textField} value={password} type="password" onChange={(event) => setPassword(event.target.value)}/>    
            </div>
            <div className="form-input">
                <label htmlFor="inputUsername">Confirm Password</label>
                <input id="inputUsername" className={classes.textField} value={passwordConfirm} type="password" onChange={(event) => setPasswordConfirm(event.target.value)}/>    
            </div>
            <Button className={classes.button + " form-button"} variant="contained" fullWidth onClick={onFormSubmit}>Finish</Button>
    </div>;
};

export default InstallPage;
