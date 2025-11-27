import { createContext, useContext, useEffect, useRef } from "react";

interface MouseManagerContextType {
    registerCallback: (
        button: MouseButton,
        callback: MouseCallback,
        options?: MouseButtonOptions
    ) => void;
    unregisterCallback: (
        button: MouseButton,
        callback: MouseCallback
    ) => void;
}

const MouseManagerContext = createContext<MouseManagerContextType>({
    registerCallback: () => { },
    unregisterCallback: () => { },
});

type MouseButton = "left" | "middle" | "right";

interface MouseButtonOptions {
    altKey?: boolean;
    ctrlKey?: boolean;
    metaKey?: boolean;
    shiftKey?: boolean;
}

type MouseCallback = (e: MouseEvent) => void;

type CallbackData = {
    callback: MouseCallback;
    options?: MouseButtonOptions;
};

export function MouseManagerProvider({ children }) {
    const callbacksRefs = useRef<Record<MouseButton, CallbackData[]>>({
        left: [],
        middle: [],
        right: [],
    });

    useEffect(() => {
        const handleClick = (e: MouseEvent) => {
            let button: MouseButton = "left";
            switch (e.button) {
                case 0:
                    button = "left";
                    break;
                case 1:
                    button = "middle";
                    break;
                case 2:
                    button = "right";
                    break;
            }

            callbacksRefs.current[button].forEach((callbackData) => {
                if (!callbackData.options) {
                    callbackData.callback(e);
                    return;
                }

                const options = callbackData.options;
                if (
                    (options.altKey ?? false) === e.altKey &&
                    (options.ctrlKey ?? false) === e.ctrlKey &&
                    (options.metaKey ?? false) === e.metaKey &&
                    (options.shiftKey ?? false) === e.shiftKey
                ) {
                    callbackData.callback(e);
                }
            });
        };

        document.addEventListener("mousedown", handleClick);
        return () => document.removeEventListener("mousedown", handleClick);
    }, []);

    function registerCallback(
        button: MouseButton,
        callback: MouseCallback,
        options?: MouseButtonOptions
    ): void {
        callbacksRefs.current[button].push({ callback, options });
    }

    function unregisterCallback(button: MouseButton, callback: MouseCallback): void {
        callbacksRefs.current[button] = callbacksRefs.current[button].filter(
            (cb) => cb.callback !== callback
        );
    }

    return (
        <MouseManagerContext.Provider
            value={{ registerCallback, unregisterCallback }}
        >
            {children}
        </MouseManagerContext.Provider>
    );
}

export const useMouseManager = () => useContext(MouseManagerContext);
