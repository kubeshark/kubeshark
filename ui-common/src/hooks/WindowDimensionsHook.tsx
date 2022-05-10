import { useState, useEffect } from 'react';

function getWindowDimensions() {
    const { innerWidth: width, innerHeight: height } = window;
    return {
        width,
        height
    };
}

export function useRequestTextByWidth(windowWidth){

    let requestText = "Request: "
    let responseText = "Response: "
    let elapsedTimeText = "Elapsed Time: "

    if (windowWidth < 1436) {
        requestText = ""
        responseText = ""
        elapsedTimeText = ""
    } else if (windowWidth < 1700) {
        requestText = "Req: "
        responseText = "Res: "
        elapsedTimeText = "ET: "
    }

    return {requestText, responseText, elapsedTimeText}
}

export default function useWindowDimensions() {
    const [windowDimensions, setWindowDimensions] = useState(getWindowDimensions());

    useEffect(() => {
        function handleResize() {
            setWindowDimensions(getWindowDimensions());
        }

        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, []);
    
    return windowDimensions;
}
