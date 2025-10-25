import { createContext } from "react";

const initialValue: useToastReturn = {
    messages: [],
    alertError: (message: string, duration?: number) => { console.log(message, duration); },
    alertSuccess: (message: string, duration?: number) => { console.log(message, duration); },
    alertInfo: (message: string, duration?: number) => { console.log(message, duration); },
};

export default createContext<useToastReturn>(initialValue);