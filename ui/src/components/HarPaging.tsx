import prevIcon from "./assets/icon-prev.svg";
import nextIcon from "./assets/icon-next.svg";
import {Box} from "@material-ui/core";
import React from "react";
import styles from './style/HarPaging.module.sass'
import numeral from 'numeral';

interface HarPagingProps {
    showPageNumber?: boolean;
}

export const HarPaging: React.FC<HarPagingProps> = ({showPageNumber=false}) => {

    return <Box className={styles.HarPaging} display='flex'>
        <img src={prevIcon} onClick={() => {
            // harStore.data.moveBack(); todo
        }} alt="back"/>
        {showPageNumber && <span className={styles.text}>
            Page <span className={styles.pageNumber}>
            {/*{numeral(harStore.data.currentPage).format(0, 0)}*/} //todo
        </span>
        </span>}
        <img src={nextIcon} onClick={() => {
            // harStore.data.moveNext(); todo
        }} alt="next"/>
    </Box>
};