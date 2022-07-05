import React from "react";
import spinner from 'spinner.svg';

export interface WithLoadingProps {
    isLoading: boolean
    loaderMargin?: number,
    loaderHeight?: number
}

const Loader = ({ loaderMargin = 20, loaderHeight = 35 }: Omit<WithLoadingProps, "isLoading">) => {
    return <div style={{ textAlign: "center", margin: loaderMargin }}>
        <img alt="spinner" src={spinner} style={{ height: loaderHeight }} />
    </div>
}

const withLoading = <P extends object>(
    Component: React.ComponentType<P>
): React.FC<P & WithLoadingProps> => ({
    isLoading,
    loaderMargin,
    loaderHeight,
    ...props
}: WithLoadingProps) => isLoading ?
            <Loader loaderMargin={loaderMargin} loaderHeight={loaderHeight} /> :
            <Component {...props as P} />;

export const LoadingWrapper: React.FC<WithLoadingProps> = ({ loaderMargin, loaderHeight, isLoading, children }) => {
    return isLoading ?
        <Loader loaderMargin={loaderMargin} loaderHeight={loaderHeight} /> :
        <React.Fragment>{children}</React.Fragment>
}

export default withLoading
