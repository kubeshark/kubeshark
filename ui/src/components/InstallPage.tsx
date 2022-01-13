import { Button } from "@material-ui/core";
import React, { useContext, useState } from "react";
import { MizuContext, Page } from "../EntApp";
import { adminUsername } from "../consts";
import Api, { FormValidationErrorType } from "../helpers/api";
import { toast } from 'react-toastify';
import LoadingOverlay from "./LoadingOverlay";
import { useCommonStyles } from "../helpers/commonStyle";

const api = Api.getInstance();

interface InstallPageProps {
    onFirstLogin: () => void;
}

export const InstallPage: React.FC<InstallPageProps> = ({onFirstLogin}) => {

    const classes = useCommonStyles();
    const [isLoading, setIsLoading] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordConfirm, setPasswordConfirm] = useState("");

    const {setPage} = useContext(MizuContext);

    const onFormSubmit = async () => {
        if (password.length < 4) {
            toast.error("Password must be at least 4 characters long");
            return;
        } else if (password !== passwordConfirm) {
            toast.error("Passwords do not match");
            return;
        }

        try {
            setIsLoading(true);
            await api.register(adminUsername, password);
            if (!await api.isAuthenticationNeeded()) {
                setPage(Page.Traffic);
                onFirstLogin();
            }
        } catch (e) {
            if (e.type === FormValidationErrorType) {
                for (const messages of Object.values(e.messages) as any[]) {
                    for (const message of messages) {
                        toast.error(message.text);
                    }
                }
            }
            console.error(e);
        } finally {
            setIsLoading(false);
        }

    }

    const handleFormOnKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter") {
            onFormSubmit();
        }
    };

    return <div className="centeredForm" onKeyPress={handleFormOnKeyPress}>
            {isLoading && <LoadingOverlay/>}
            <div className="form-title left-text">Setup</div>
            <span className="form-subtitle">Welcome to Mizu, please set up the admin user to continue</span>
            <div className="form-input">
                <label htmlFor="inputUsername">Username</label>
                <input id="inputUsername" className={classes.textField} value={adminUsername} disabled={true} />
            </div>
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
