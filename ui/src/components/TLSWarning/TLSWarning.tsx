import {Snackbar} from "@material-ui/core";
import MuiAlert from "@material-ui/lab/Alert";
import React, {useEffect} from "react";
import Api from "../../helpers/api";
import './TLSWarning.sass';

const api = new Api();

interface TLSWarningProps {
    showTLSWarning: boolean
    setShowTLSWarning: (show: boolean) => void
    addressesWithTLS: Set<string>
    setAddressesWithTLS: (addresses: Set<string>) => void
    userDismissedTLSWarning: boolean
    setUserDismissedTLSWarning: (flag: boolean) => void
}

export const TLSWarning: React.FC<TLSWarningProps>  = ({showTLSWarning, setShowTLSWarning, addressesWithTLS, setAddressesWithTLS, userDismissedTLSWarning, setUserDismissedTLSWarning}) => {

    useEffect(() => {
        (async () => {
            try {
                const recentTLSLinks = await api.getRecentTLSLinks();
                if (recentTLSLinks?.length > 0) {
                    setAddressesWithTLS(new Set(recentTLSLinks));
                    setShowTLSWarning(true);
                }
            } catch (e) {
                console.error(e);
            }
        })();
    }, []);

    return (<Snackbar open={showTLSWarning && !userDismissedTLSWarning}>
        <MuiAlert classes={{filledWarning: 'customWarningStyle'}} elevation={6} variant="filled"
                  onClose={() => setUserDismissedTLSWarning(true)} severity="warning">
            Mizu is detecting TLS traffic, this type of traffic will not be displayed.
            {addressesWithTLS.size > 0 &&
            <ul className="httpsDomains"> {Array.from(addressesWithTLS, address => <li>{address}</li>)} </ul>}
        </MuiAlert>
    </Snackbar>);
}
