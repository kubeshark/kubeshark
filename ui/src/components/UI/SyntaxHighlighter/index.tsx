import React from 'react';
import {Prism as SyntaxHighlighterContainer} from 'react-syntax-highlighter';
import {highlighterStyle} from './highlighterStyle'
import './index.scss';

interface Props {
    code: string;
    style?: any;
    showLineNumbers?: boolean;
    className?: string;
    language?: string;
    isWrapped?: boolean;
}

export const SyntaxHighlighter: React.FC<Props> = ({
                                                code,
                                                style = highlighterStyle,
                                                showLineNumbers = true,
                                                className,
                                                language = 'python',
                                                isWrapped = false,
                                            }) => {
    return <div className={`highlighterContainer ${className ? className : ''} ${isWrapped ? 'wrapped' : ''}`}>
        <SyntaxHighlighterContainer language={language} style={style} showLineNumbers={showLineNumbers}>
            {code ?? ""}
        </SyntaxHighlighterContainer>
    </div>;
};

export default SyntaxHighlighter;
