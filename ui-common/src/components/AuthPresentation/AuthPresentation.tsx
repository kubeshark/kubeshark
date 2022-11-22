import React, {useEffect, useState} from "react";
import Api from "../../helpers/api";
import './AuthPresentation.sass';

const api = Api.getInstance();

export const AuthPresentation = () => {

    const [statusAuth, setStatusAuth] = useState(null);

    useEffect(() => {
        (async () => {
            try {
                const auth = await api.getAuthStatus();
                setStatusAuth(auth);
            } catch (e) {
                console.error(e);
            }
        })();
    }, []);

    return <>
        {statusAuth?.email && <div className="authPresentationContainer">
                <div>
                    <div className="authEmail">{statusAuth.email}</div>
                    <div className="authModel">{statusAuth.model}</div>
                </div>
            </div>}
    </>;
}
