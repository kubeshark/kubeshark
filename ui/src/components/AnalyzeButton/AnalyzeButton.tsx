import {Button} from "@material-ui/core";
import React from "react";
import {UI} from "@up9/mizu-common";
import logo_up9 from "../assets/logo_up9.svg";
import {makeStyles} from "@material-ui/core/styles";

const useStyles = makeStyles(() => ({
    tooltip: {
        backgroundColor: "#3868dc",
        color: "white",
        fontSize: 13,
    },
}));

interface AnalyseButtonProps {
    analyzeStatus: any
}

export const AnalyzeButton: React.FC<AnalyseButtonProps>  = ({analyzeStatus}) => {

    const classes = useStyles();

    const analysisMessage = analyzeStatus?.isRemoteReady ?
        <span>
            <table>
                <tr>
                    <td>Status</td>
                    <td><b>Available</b></td>
                </tr>
                <tr>
                    <td>Messages</td>
                    <td><b>{analyzeStatus?.sentCount}</b></td>
                </tr>
            </table>
        </span> :
        analyzeStatus?.sentCount > 0 ?
            <span>
                <table>
                    <tr>
                        <td>Status</td>
                        <td><b>Processing</b></td>
                    </tr>
                    <tr>
                        <td>Messages</td>
                        <td><b>{analyzeStatus?.sentCount}</b></td>
                    </tr>
                    <tr>
                        <td colSpan={2}> Please allow a few minutes for the analysis to complete</td>
                    </tr>
                </table>
            </span> :
            <span>
                <table>
                    <tr>
                        <td>Status</td>
                        <td><b>Waiting for traffic</b></td>
                    </tr>
                    <tr>
                        <td>Messages</td>
                        <td><b>{analyzeStatus?.sentCount}</b></td>
                    </tr>
                </table>
            </span>

    return ( <div>
        <UI.Tooltip title={analysisMessage} isSimple classes={classes}>
            <div>
                <Button
                    style={{fontFamily: "system-ui",
                        fontWeight: 600,
                        fontSize: 12,
                        padding: 8}}
                    size={"small"}
                    variant="contained"
                    color="primary"
                    startIcon={<img style={{height: 24, maxHeight: "none", maxWidth: "none"}} src={logo_up9} alt={"up9"}/>}
                    disabled={!analyzeStatus?.isRemoteReady}
                    onClick={() => {
                        window.open(analyzeStatus?.remoteUrl)
                    }}>
                    Analysis
                </Button>
            </div>
        </UI.Tooltip>
    </div>);
}
