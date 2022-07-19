import React from 'react';
import { useEffect } from 'react';
import { useFilePicker } from 'use-file-picker';
import { FileContent } from 'use-file-picker/dist/interfaces';

interface IFilePickerProps {
    onLoadingComplete: (file: FileContent) => void;
    elem: any
}

const FilePicker = ({ elem, onLoadingComplete }: IFilePickerProps) => {
    const [openFileSelector, { filesContent }] = useFilePicker({
        accept: ['.json'],
        limitFilesConfig: { max: 1 },
        maxFileSize: 1
    });

    const onFileSelectorClick = (e) => {
        e.preventDefault();
        e.stopPropagation();
        openFileSelector();
    }

    useEffect(() => {
        filesContent.length && onLoadingComplete(filesContent[0])
    }, [filesContent, onLoadingComplete]);

    return (<React.Fragment>
        {React.cloneElement(elem, { onClick: onFileSelectorClick })}
    </React.Fragment>)
}

export default FilePicker;
