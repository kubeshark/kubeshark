import React from "react";

import styles from './style/Resizeable.module.sass'

export interface Props {
    children
}

const Resizeable: React.FC<Props> = ({ children }) => {

    const [initialPos, setInitialPos] = React.useState(null || Number);
    const [initialSize, setInitialSize] = React.useState(null);

    const initial = (e) => {
        let resizable = document.getElementById('Resizable');
        setInitialPos(e.clientX);
        setInitialSize(resizable.offsetWidth);

    }

    const resize = (e: any) => {
        let resizable = document.getElementById('Resizable');
        resizable.style.width = `${parseInt(initialSize) + parseInt(e.clientX) - initialPos}px`
    }

    return (
        <React.Fragment>
            <div className={styles.Block}>
                <div id='Resizable' className={styles.Resizable}>
                    {children}
                </div>
                <div id='Draggable' className={styles.Draggable}
                    draggable='true'
                    onDragStart={initial}
                    onDrag={resize}
                />
            </div>
        </React.Fragment>
    );
};

export default Resizeable;