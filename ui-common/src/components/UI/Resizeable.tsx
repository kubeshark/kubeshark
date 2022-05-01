import React, { useRef, useState } from "react";

import styles from './style/Resizeable.module.sass'

export interface Props {
    children
    minWidth: number
    maxWidth?: number
}

const Resizeable: React.FC<Props> = ({ children, minWidth, maxWidth }) => {
    const resizeble = useRef(null)
    let mousePos = { x: 0, y: 0 }
    let elementDimention = { w: 0, h: 0 }
    let isPressed = false
    const [elemWidth, setElemWidth] = useState(resizeble?.current?.style?.width)

    const mouseDownHandler = function (e) {
        // Get the current mouse position
        mousePos = { x: e.clientX, y: e.clientY }
        isPressed = true

        // Calculate the dimension of element
        const styles = resizeble.current.getBoundingClientRect();
        elementDimention = { w: parseInt(styles.width, 10), h: parseInt(styles.height, 10) }
        // Attach the listeners to `document`
        window.addEventListener('mousemove', mouseMoveHandler);
        window.addEventListener('mouseup', mouseUpHandler);
    };


    const mouseMoveHandler = function (e) {
        if (isPressed) {
            // How far the mouse has been moved
            const dx = e.clientX - mousePos.x;
            const widthEl = elementDimention.w + dx

            if (widthEl >= minWidth)
                // Adjust the dimension of element
                setElemWidth(widthEl)
        }
    };

    const mouseUpHandler = function () {
        window.removeEventListener('mousemove', mouseMoveHandler);
        window.removeEventListener('mouseup', mouseUpHandler);
        isPressed = false
    };

    return (
        <div className={styles.resizable} ref={resizeble} style={{ width: elemWidth, maxWidth: maxWidth }}>
            {children}
            <div className={`${styles.resizer} ${styles.resizerRight}`} onMouseDown={mouseDownHandler}></div>
            {/* <div className={`${styles.resizer} ${styles.resizerB}`} onMouseDown={mouseDownHandler}></div> -- FutureUse*/}
        </div>
    );
};

export default Resizeable;