import React from "react";
import makeStyles from '@mui/styles/makeStyles';
import { Autocomplete } from "@mui/material";
import { Checkbox, TextField } from "@mui/material";
import CheckBoxOutlineBlankIcon from '@mui/icons-material/CheckBoxOutlineBlank';
import CheckBoxIcon from '@mui/icons-material/CheckBox';
import DefaultIconDown from "DefaultIconDown.svg";
import styles from "./SearchableDropdown.module.sass";

interface SearchableDropdownProps {
    options: string[],
    selectedValues?: any,
    onChange: (string) => void,
    isLoading?: boolean,
    label?: string,
    multiple?: boolean,
    inputWidth?: string
    freeSolo?: boolean
}

const useStyles = makeStyles(() => ({

    inputRoot: {
        padding: "8px 4px 8px 12px !important",
        borderRadius: "9px !important"
    },
    input: {
        padding: "0px 33px 0px 0px !important",
        height: 18,
        fontWeight: "normal",
        fontFamily: "Source Sans Pro, Lucida Grande, Tahoma, sans-serif"
    },
    root: {
        "& .MuiFormLabel-root": {
            fontFamily: "Source Sans Pro, Lucida Grande, Tahoma, sans-serif"
        }
    }
}));


const SearchableDropdown: React.FC<SearchableDropdownProps> = ({ options, selectedValues, onChange, isLoading, label, multiple, inputWidth, freeSolo }) => {

    const classes = useStyles();

    return <Autocomplete
        freeSolo={freeSolo}
        multiple={multiple}
        loading={isLoading}
        classes={classes}
        options={options ? options : []}
        disableCloseOnSelect={multiple}
        fullWidth={false}
        disableClearable
        value={selectedValues ? selectedValues : (multiple ? [] : null)}
        getOptionLabel={(option) => option}
        onChange={(event, val) => onChange(val)}
        size={"small"}
        popupIcon={<img style={{ padding: 7 }} alt="iconDown" src={DefaultIconDown} />}
        renderOption={(props, option, { selected }) => (
            <li {...props}>
            <div id={`option-${option}`} className={styles.optionItem} key={option}>
                {multiple && <Checkbox
                    icon={<CheckBoxOutlineBlankIcon fontSize="small" />}
                    checkedIcon={<CheckBoxIcon fontSize="small" />}
                    style={{ marginRight: 8 }}
                    checked={selected}
                />}
                <div title={option} className={styles.title}>{option}</div>
            </div>
            </li>
        )}
        renderTags={() => <div className={styles.optionListItem}>
            <div title={selectedValues?.length > 0 ? `${selectedValues[0]} (+${selectedValues.length - 1})` : ""} className={styles.optionListItemTitle}>
                {selectedValues?.length > 0 ? `${selectedValues[0]}` : ""}
            </div>
            {selectedValues?.length > 1 && <div style={{ marginLeft: 5 }}>{`(+${selectedValues.length - 1})`}</div>}
        </div>}
        renderInput={(params) => <TextField
            onChange={(e) => freeSolo && onChange(e.target.value)}
            variant="outlined" label={label}
            className={`${classes.root} searchableDropdown`}
            style={{ backgroundColor: "transparent" }}
            {...params}
        />
        }
    />
};

export default SearchableDropdown;
