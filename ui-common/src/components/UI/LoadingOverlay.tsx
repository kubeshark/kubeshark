import React, {useEffect, useState} from "react";
import style from '../style/LoadingOverlay.module.sass';

const SpinnerShowDelayMs = 350;

interface LoadingOverlayProps {
    delay?: number
}

const LoadingOverlay: React.FC<LoadingOverlayProps> = ({delay}) => {

    const [isVisible, setIsVisible] = useState(false);

    // @ts-ignore
    useEffect(() => {
        let isRelevant = true;

        setTimeout(() => {
            if(isRelevant)
                setIsVisible(true);
        }, delay || SpinnerShowDelayMs);

        return () => isRelevant = false;
    }, [delay]);

    return <div className={style.loadingOverlayContainer} hidden={!isVisible}>
        <div className={style.loadingOverlaySpinner}/>
    </div>
};

export default LoadingOverlay;
