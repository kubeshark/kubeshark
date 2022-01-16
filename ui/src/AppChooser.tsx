import React from 'react';

const AppChooser = () => {

    let mainComponent;
    if(process.env.REACT_APP_OVERRIDE_IS_ENTERPRISE === "true") {
        mainComponent = React.lazy(() => import('./EntApp'));
    } else {
        mainComponent = <div>blabla</div>
    }

    return (
       <>{mainComponent}</>
    );
}

export default AppChooser;
