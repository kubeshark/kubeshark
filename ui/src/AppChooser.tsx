import React, {Suspense} from 'react';
import LoadingOverlay from "./components/LoadingOverlay";

const AppChooser = () => {

    let MainComponent;
    if(process.env.REACT_APP_OVERRIDE_IS_ENTERPRISE === "true") {
        MainComponent = React.lazy(() => import('./EntApp'));
    } else {
        MainComponent = React.lazy(() => import('./App'));
    }

    return <Suspense fallback={<LoadingOverlay/>}>
        <MainComponent/>
    </Suspense>;
}

export default AppChooser;
