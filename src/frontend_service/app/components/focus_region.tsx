import React, { type CSSProperties } from "react";
import { useFocusManager } from "./focus_manager";

interface FocusRegionProps {
    id: string;
    children: React.ReactNode;
    className?: string;
    onClick?: React.MouseEventHandler<HTMLDivElement>;
    style?: CSSProperties;
};

export function FocusRegion({ id, children, className, onClick, style }: FocusRegionProps) {
    const { activeRegion, setActiveRegion } = useFocusManager();
    const isActive = activeRegion === id;

    function onClickHandler(event) {
        setActiveRegion(id);
        if (onClick) {
            onClick(event);
        }
    }

    if (!isActive) { return <></>; }
    return (
        <div
            id={id}
            className={className}
            onClick={onClickHandler}
            style={style}
            data-focus-region
        >
            {children}
        </div>
    );
};
