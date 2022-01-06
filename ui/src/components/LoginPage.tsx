import { Button, TextField } from "@material-ui/core";
import React, { useContext, useState } from "react";
import { toast } from "react-toastify";
import { MizuContext, Page } from "../EntApp";
import Api from "../helpers/api";
import LoadingOverlay from "./LoadingOverlay";

const api = Api.getInstance();

const LoginPage: React.FC = () => {

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
            <p>Welcome to Mizu, please login to continue</p>
            <TextField className="form-input" label="Username" variant="standard" fullWidth value={username} onChange={e => setUsername(e.target.value)} />
            <TextField className="form-input" label="Password" variant="standard" type="password" fullWidth value={password} onChange={e => setPassword(e.target.value)} />
            <Button className="form-button" variant="contained" fullWidth onClick={onFormSubmit}>Login</Button>
    </div>;
};

export default LoginPage
