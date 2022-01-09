import { Button, TextField } from "@material-ui/core";
import React, { useContext, useState } from "react";
import { toast } from "react-toastify";
import { MizuContext, Page } from "../EntApp";
import Api from "../helpers/api";
import { useCommonStyles } from "../helpers/commonStyle";
import LoadingOverlay from "./LoadingOverlay";

const api = Api.getInstance();

const LoginPage: React.FC = () => {

    const classes = useCommonStyles();
    const [isLoading, setIsLoading] = useState(false);
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");

    const {setPage} = useContext(MizuContext);

    const onFormSubmit = async () => {
        setIsLoading(true);

        try {
            await api.login(username, password);
            if (!await api.isAuthenticationNeeded()) {
                setPage(Page.Traffic);
            } else {
                toast.error("Invalid credentials");
            }
        } catch (e) {
            toast.error("Invalid credentials");
            console.error(e);
        } finally {
            setIsLoading(false);
        }

    }


    return <div className="centeredForm">
            {isLoading && <LoadingOverlay/>}
            <div className="form-title left-text">Login</div>
            <div className="form-input">
                <label htmlFor="inputUsername">Username</label>
                <input id="inputUsername" className={classes.textField} value={username} onChange={(event) => setUsername(event.target.value)}/>    
            </div>
            <div className="form-input">
                <label htmlFor="inputPassword">Password</label>
                <input id="inputPassword" className={classes.textField} value={password} type="password" onChange={(event) => setPassword(event.target.value)}/>
            </div>
            <Button className={classes.button + " form-button"} variant="contained" fullWidth onClick={onFormSubmit}>Log in</Button>
            
    </div>;
};

export default LoginPage
