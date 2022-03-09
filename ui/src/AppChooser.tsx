import EntApp from './EntApp';

const AppChooser = () => {

    // let MainComponent;
    // if (process.env.REACT_APP_OVERRIDE_IS_ENTERPRISE === "true") {
    //     MainComponent = React.lazy(() => import('./EntApp'));
    // } else {
    //     MainComponent = React.lazy(() => import('./App'));
    // }
 // return <Suspense fallback={<UI.LoadingOverlay/>}>
    //     <MainComponent/>
    // </Suspense>;

    return <EntApp/>
   
}

export default AppChooser;