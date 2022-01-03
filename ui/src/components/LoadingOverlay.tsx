import React, {useEffect, useState} from "react";
import './style/LoadingOverlay.sass';

const SpinnerShowDelayMs = 350;

interface LoadingOverlayProps {
    delay?: number
}

const LoadingOverlay: React.FC<LoadingOverlayProps> = ({delay}) => {

    const [isVisible, setIsVisible] = useState(false);

    //@ts-ignore
    useEffect(() => {
        let isRelevant = true;

        setTimeout(() => {
            if (isRelevant)
                setIsVisible(true);
        }, delay ?? SpinnerShowDelayMs);

        return () => isRelevant = false;
    }, []);

    return <div className="loading-overlay-container" hidden={!isVisible}>
        <div className="loading-overlay-spinner"/>
    </div>
};

export default LoadingOverlay;
