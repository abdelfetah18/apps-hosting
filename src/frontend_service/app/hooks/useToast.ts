import { useState } from "react";

const DEFAULT_DURATION = 3000;

export default function useToast(): useToastReturn {
    let MESSAGE_ID = 0;
    const [messages, setMessages] = useState<ToastMessage[]>([]);

    const alert = (message: string, type: ToastMessageType, duration?: number): void => {
        let id = MESSAGE_ID;
        setMessages(state => [...state, { id: id, message, type }]);
        setTimeout(() => {
            setMessages(state => {
                let msgs = [...state];
                return msgs.filter(msg => msg.id != id);
            });
        }, duration || DEFAULT_DURATION);
        MESSAGE_ID++;
        console.log(messages);
    }

    const alertError = (message: string, duration?: number) => {
        alert(message, "error", duration);
    }

    const alertInfo = (message: string, duration?: number) => {
        alert(message, "info", duration);
    }

    const alertSuccess = (message: string, duration?: number) => {
        alert(message, "success", duration);
    }

    return {
        messages,
        alertError,
        alertInfo,
        alertSuccess
    };
}