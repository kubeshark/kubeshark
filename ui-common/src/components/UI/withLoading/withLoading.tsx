import React from "react";
import spinner from 'spinner.svg';

export interface WithLoadingProps {
    isLoading: boolean
    loaderMargin?: number,
    loaderHeight?: number
}

const withLoading = <P extends object>(
    Component: React.ComponentType<P>
): React.FC<P & WithLoadingProps> => ({
    isLoading,
    loaderMargin = 20,
    loaderHeight = 35,
    ...props
}: WithLoadingProps) => isLoading ? <div style={{ textAlign: "center", margin: 20 }}>
    <img alt="spinner" src={spinner} style={{ height: 35 }} />
</div> : <Component {...props as P} />;


export default withLoading
