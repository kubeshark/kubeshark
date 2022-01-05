import React, {useEffect, useState} from "react";
import './style/LoadingOverlay.sass';

const SpinnerShowDelayMs = 350;

interface LoadingOverlayProps {
    delay?: number
}

const LoadingOverlay: React.FC<LoadingOverlayProps> = ({delay}) => {

    const [isVisible, setIsVisible] = useState(false);

    useEffect(() => {
        setTimeout(() => {
            setIsVisible(true);
        }, delay ?? SpinnerShowDelayMs);
    }, []);

    return <div className="loading-overlay-container" hidden={!isVisible}>
        <div className="loading-overlay-spinner"/>
    </div>
};

export default LoadingOverlay;
