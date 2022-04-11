import React from "react";

import styles from './style/Resizeable.module.sass'

export interface Props {
    children
}

const Resizeable: React.FC<Props> = ({ children }) => {

    const [initialPos, setInitialPos] = React.useState(null || Number);
    const [initialPosDraggable, setInitialPosDraggable] = React.useState(null || Number);
    const [initialSize, setInitialSize] = React.useState(null);

    const initial = (e) => {
        let resizable = document.getElementById('Resizable');
        setInitialPos(e.clientX);
        setInitialPosDraggable(e.clientX)
        setInitialSize(resizable.offsetWidth);

    }

    const draOver = (e) => {
        e.preventDefault();
    }

    const resize = (e: any) => {
        e.preventDefault()
        let resizable = document.getElementById('Resizable');
        resizable.style.width = `${parseInt(initialSize) + parseInt(e.clientX) - initialPos}px`
        let draggable = document.getElementById('Draggable');
        draggable.style.left = `${parseInt(e.clientX) - initialPosDraggable}px`

    }

    return (
        <React.Fragment>
            <div className={styles.Block}>
                <div id='Resizable' className={styles.Resizable} draggable={false}>
                    {children}
                </div>
                <div id='Draggable' className={styles.Draggable}
                    onDragOver={draOver}
                    onDragStart={initial}
                    onDrag={resize}
                    draggable
                />
            </div>
        </React.Fragment>
    );
};

export default Resizeable;