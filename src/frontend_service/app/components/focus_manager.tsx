import { createContext, useContext, useEffect, useState } from "react";
import { useMouseManager } from "./mouse_manager";


interface FocusManagerContextType {
    activeRegion: string;
    setActiveRegion: React.Dispatch<React.SetStateAction<string>>;
};

const FocusManagerContext = createContext<FocusManagerContextType>({
    activeRegion: "",
    setActiveRegion: () => { },
});

export function FocusManagerProvider({ children }) {
    const [activeRegion, setActiveRegion] = useState<string>("");
    const { registerCallback, unregisterCallback } = useMouseManager();

    useEffect(() => {
        const handleClick = (e: MouseEvent) => {
            const clickedInRegion = (e.target as HTMLElement).closest("[data-focus-region]");
            if (!clickedInRegion) {
                setActiveRegion("");
            }
        };
        registerCallback("left", handleClick);
        return () => unregisterCallback("left", handleClick);
    }, []);

    return (
        <FocusManagerContext.Provider value={{ activeRegion, setActiveRegion }}>
            {children}
        </FocusManagerContext.Provider>
    );
};

export const useFocusManager = () => useContext(FocusManagerContext);
