import React from "react";
import AnsiToHtml from "ansi-to-html";

export default function AnsiLogs({ logs }: { logs: string }) {
    const ansiConverter = new AnsiToHtml({
        fg: "#ccc",
        bg: "#111",
        newline: true,
        escapeXML: true,
        stream: false,
    });

    const html = ansiConverter.toHtml(logs);

    return (
        <div
            style={{
                background: "#111",
                color: "#ccc",
                padding: "20px",
                fontFamily: "monospace",
                whiteSpace: "pre-wrap",
                height: "100%",
                overflow: "auto",
            }}
            dangerouslySetInnerHTML={{ __html: html }}
        />
    );
};
